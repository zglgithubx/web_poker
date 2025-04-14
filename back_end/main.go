package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"regexp"
	"sort"
	"strings"

	"github.com/gorilla/websocket"
	ds "mq/data_struct"
	ut "mq/utils"
)

func main() {
	gs := NewGameServer()
	http.HandleFunc("/ws", gs.HandleConnections)
	log.Fatal(http.ListenAndServe(":8080", nil))
}

type GameServer ds.GameServer

var logger *log.Logger
var upgrader = websocket.Upgrader{
	CheckOrigin:     func(r *http.Request) bool { return true },
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

func NewGameServer() *GameServer {
	gs := &GameServer{
		Rooms: make(map[string]*ds.Room),
		Lobby: make(chan *ds.Player),
		Msg:   make(chan *ds.TransferMessage),
	}
	//匹配玩家
	go gs.MatchPlayers()
	//处理消息
	go gs.HandleMessage()
	logger = log.New(os.Stdout, "INFO: ", log.Ldate|log.Ltime|log.Lshortfile)
	return gs
}
func (gs *GameServer) HandleConnections(w http.ResponseWriter, r *http.Request) {
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Fatal(err)
	}
	connId := ut.GetUUID()
	player := &ds.Player{Conn: ws, ID: connId}
	gs.Lobby <- player
	defer func(ws *websocket.Conn) {
		err := ws.Close()
		if err != nil {
			logger.Println(err.Error())
		}
	}(ws)
	for {
		msgType, msg, err := ws.ReadMessage()
		curRoomPlayers := gs.Rooms[player.RooId].Players
		curRoomPlayersIndexIndex := gs.Rooms[player.RooId].PlayersIndex
		logger.Println("当前用户信息:", player)
		if err != nil || msgType == -1 {
			delete(curRoomPlayers, connId)
			delete(curRoomPlayersIndexIndex, connId)
			gs.Rooms[player.RooId].Status = "waiting"
			logger.Printf("当前玩家：%v,断开连接", player)
			break
		}
		message := new(ds.TransferMessage)
		unmarshelError := json.Unmarshal(msg, message)
		if unmarshelError != nil {
			logger.Printf("反序列化消息失败:%v", unmarshelError)
			continue
		}
		fmt.Printf("%s 发送: %s,消息类型：%d\n", ws.RemoteAddr(), message, msgType)
		if message.Type == "init" {
			data := message.Data
			playerID, ok1 := data["playerId"].(string)
			playerName, ok2 := data["playerName"].(string)
			if !ok1 || !ok2 {
				logger.Println("Invalid player data")
				continue
			}
			curRoomPlayers[connId].Name = playerName
			curRoomPlayers[connId].UserId = playerID
		} else {
			gs.Msg <- message
		}
	}
}
func (gs *GameServer) MatchPlayers() {
	for {
		//匹配玩家
		player := <-gs.Lobby
		gs.Mutex.Lock()
		hasRoom := false
		responseMessage := ds.TransferMessage{
			Data: make(map[string]any),
		}
		responseMessage.Type = "waiting"
		responseMessage.Data["connId"] = player.ID
		for _, room := range gs.Rooms {
			if len(room.Players) < 3 && room.Status == "waiting" {
				player.RooId = room.ID
				room.Players[player.ID] = player
				room.PlayersIndex[player.ID] = len(room.Players) + 1
				// 房间满了，开始游戏
				if len(room.Players) == 3 {
					room.Status = "playing"
					responseMessage.Type = "playing"
					// 发牌
					room.InitDeck()
					//开始游戏
					go room.StartGame()
				}
				gs.Mutex.Unlock()
				hasRoom = true
				ut.ResponseResult(player.Conn, responseMessage)
				break
			}
		}
		if !hasRoom {
			// 创建新房间
			newRoom := &ds.Room{
				ID:                ut.GenerateRoomID(),
				Status:            "waiting",
				PlayersIndex:      make(map[string]int),
				Players:           make(map[string]*ds.Player),
				LastUserId:        "0",
				LastCards:         make([]ds.Card, 0),
				LastIsPass:        false,
				CurSendCardUserId: "0",
			}
			player.RooId = newRoom.ID
			newRoom.Players[player.ID] = player
			newRoom.PlayersIndex[player.ID] = 1
			gs.Rooms[newRoom.ID] = newRoom
			gs.Mutex.Unlock()
			ut.ResponseResult(player.Conn, responseMessage)
		}
	}
}
func (gs *GameServer) HandleMessage() {
	for {
		message := <-gs.Msg
		fmt.Printf("开始处理消息: %v\n", message)
		// 假设消息中包含 playerId
		var data map[string]any = message.Data
		connId, ok := data["connId"].(string)
		if !ok {
			fmt.Println("消息中缺少 playerId")
			continue
		}
		curRoomPlayers := gs.Rooms[connId].Players

		gs.Mutex.RLock()
		player, exists := curRoomPlayers[connId]
		gs.Mutex.RUnlock()

		if !exists {
			fmt.Printf("未找到 connId 为 %s 的玩家\n", connId)
			continue
		}

		// 现在可以使用 player 对象
		fmt.Printf("当前玩家信息: %+v\n", player)

		playerSendCardStr, ok := data["sendCard"].(string)
		if !ok {
			fmt.Println("消息中缺少发牌信息！")
			continue
		}
		sendCards := ds.Cards{}
		err := json.Unmarshal([]byte(playerSendCardStr), &sendCards)
		if err != nil {
			fmt.Println("消息反序列化失败！")
			continue
		}
		responseMsg := ds.ResponseMessage{}
		curRoom := gs.Rooms[player.RooId]
		if curRoom.LastUserId != "0" && player.ID != curRoom.CurSendCardUserId {
			//出牌人不是该玩家
			responseMsg.SendCardStatus = -2
			responseMsg.Describe = "不该你出牌！"
			ut.ResponseResult(player.Conn, responseMsg)
			continue
		}
		//处理打牌逻辑
		//组装响应其他玩家的消息体
		sendCardLen := len(sendCards)
		receptionPlayers := []ds.Player{}
		curPlayIndex := gs.Rooms[player.RooId].PlayersIndex[connId]

		if curPlayIndex == 2 {
			responseMsg.NextUserId = curRoom.Players[connId].ID
		} else {
			// responseMsg.NextUserId = curRoom.Players[curPlayIndex+1].ID
		}
		curRoom.CurSendCardUserId = responseMsg.NextUserId
		//当前玩家 Pass
		if sendCardLen == 0 {
			curRoom.LastIsPass = true
			responseMsg.SendCardStatus = 0
			for _, otherPlay := range receptionPlayers {
				ut.ResponseResult(otherPlay.Conn, responseMsg)
			}
			continue
		}
		sort.Sort(sendCards)
		ranks := []string{"A", "2", "3", "4", "5", "6", "7", "8", "9", "I", "J", "Q", "K"}
		mapRanks := make(map[string]int, 0)
		for index, rank := range ranks {
			mapRanks[rank] = index + 1
		}
		userId := player.ID
		//上一个出牌的id为空，则为第一次出牌
		if curRoom.LastUserId == "0" {
			curRoom.LastUserId = userId
			curRoom.LastCards = sendCards
			curRoom.LastIsPass = false
		} else {
			//当前玩家出牌之后，其他玩家pass
			if curRoom.LastIsPass && curRoom.LastUserId == userId {
				curRoom.LastUserId = userId
				curRoom.LastCards = sendCards
				curRoom.LastIsPass = false
			} else {
				//开始比较当前出的牌和上一次出牌大小
				//首先比较牌的数量是否一致
				curLen := len(sendCards)
				//判断是否是王炸
				if curLen == 2 && sendCards[0].Value == 14 && sendCards[1].Value == 15 {
					curRoom.LastUserId = userId
					curRoom.LastCards = sendCards
					curRoom.LastIsPass = false
				} else {
					//格式不对，重新出牌
					if curLen != len(curRoom.LastCards) {
						responseMsg.SendCardStatus = -1
						responseMsg.Describe = "格式不对，请重新出牌！"
						ut.ResponseResult(player.Conn, responseMsg)
						continue
					}
					var lastStr, curStr string
					var lastValue, curValue int
					for _, curCard := range sendCards {
						curStr += curCard.Rank
						curValue += curCard.Value
					}
					for _, curCard := range curRoom.LastCards {
						lastStr += curCard.Rank
						lastValue += curCard.Value
					}
					//判断是否是特殊情况
					if curLen > 3 && curLen < 7 {
						var allChar string = "A23456789IJQK"
						//匹配是否是四带一或者四带二
						re := regexp.MustCompile(fmt.Sprintf("[%s]{4}", allChar))
						//匹配是否是三带一或者三带二
						re1 := regexp.MustCompile(fmt.Sprintf("[%s]{3}", allChar))
						reFindCurStr := re.FindString(curStr)
						reFindLastStr := re.FindString(lastStr)
						re1FindCurStr := re1.FindString(curStr)
						re1FindLastStr := re1.FindString(lastStr)
						if strings.EqualFold(reFindCurStr, reFindLastStr) {
							curValue = mapRanks[string(reFindCurStr[0])]
							lastValue = mapRanks[string(reFindLastStr[0])]
						} else if strings.EqualFold(re1FindCurStr, re1FindLastStr) {
							curValue = mapRanks[string(re1FindCurStr[0])]
							lastValue = mapRanks[string(re1FindLastStr[0])]
						}
					}
					//成功出牌
					if curValue > lastValue {
						//记录上一次出牌和上一次出牌人
						curRoom.LastUserId = userId
						curRoom.LastCards = sendCards
						curRoom.LastIsPass = false
						responseMsg.SendCardStatus = 2
					} else {
						//压不起，重新出牌
						responseMsg.SendCardStatus = 1
						responseMsg.Describe = "压不过上一次牌，请重新出牌"
						ut.ResponseResult(player.Conn, responseMsg)
						continue
					}
				}
			}
		}
		//组装结果消息，发送给其他两人
		responseMsg.CurSendCard = sendCards
		responseMsg.SendCardStatus = 2
		responseMsg.Describe = "继续出牌！"
		for _, otherPlay := range receptionPlayers {
			ut.ResponseResult(otherPlay.Conn, responseMsg)
		}
	}
}
