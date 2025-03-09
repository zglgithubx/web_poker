package main

import (
	"fmt"
	"log"
	"net/http"
	"regexp"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin:     func(r *http.Request) bool { return true },
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}
var userNumber = 1

type GameServer struct {
	Rooms map[string]*Room
	Lobby chan *Player
	Msg   chan *Message
	mutex sync.RWMutex
}
type Room struct {
	ID      string
	Players []*Player
	Deck    []Card
	mutex   sync.Mutex
	Status  string // waiting/playing
	LastCards Cards
	LastUserId int
	LastIsPass bool
	CurSendCardUserId int
}
type Player struct {
	Conn  *websocket.Conn
	RooId string
	ID    int
	Name  string
}
type Card struct {
	Suit  string
	Rank  string
	Value int
}
type Cards []Card
func(c Cards) Len()int{
	return len(c)
}
func(c Cards) Less(i,j int)bool{
	return c[i].Value<c[j].Value
}
func(c Cards)Swap(i,j int){
	c[i], c[j] = c[j], c[i]
}
type Message struct {
	SendCard Cards `json:"sendCard"`
}
type ResponseMessage struct{
	//-2: 不该你出牌，-1: 格式不对重新出牌，0 : 不出，1：压不过，重新出牌，2：出牌
	SendCardStatus int `json:"sendCardStatus"`
	CurSendCard Cards `json:"sendCard"`
	//下一个出牌的用户id
	NextUserId int `json:"nextUserId"`
	//当前响应消息的描述
	Describe string `json:"describe"`
}
func (r *Room) InitDeck() {
	
	r.mutex.Lock()
	defer r.mutex.Unlock()
	r.Deck = nil
	suits := []string{"♠", "♥", "♦", "♣"}
	ranks := []string{"A", "2", "3", "4", "5", "6", "7", "8", "9", "I", "J", "Q", "K"}
	for value, rank := range ranks {
		for _, suit := range suits {
			r.Deck = append(r.Deck, Card{
				Suit:  suit,
				Rank:  rank,
				Value: value + 1,
			})
		}
	}
	//大王
	r.Deck = append(r.Deck, Card{
		Suit:  "大",
		Rank:  "王",
		Value: 15,
	})
	//小王
	r.Deck = append(r.Deck, Card{
		Suit:  "小",
		Rank:  "王",
		Value: 14,
	})
}

func NewGameServer() *GameServer {
	gs := &GameServer{
		Rooms: make(map[string]*Room),
		Lobby: make(chan *Player),
		Msg:   make(chan *Message),
	}
	//计算玩家
	go gs.MatchPlayers()
	return gs
}

func (gs *GameServer) MatchPlayers() {
	for {
		//匹配玩家
		player := <-gs.Lobby
		gs.mutex.Lock()
		hasRoom := false
		for _, room := range gs.Rooms {
			if len(room.Players) < 3 && room.Status == "waiting" {
				player.RooId = room.ID
				room.Players = append(room.Players, player)
				player.Conn.WriteJSON(player)
				// 房间满了，开始游戏
				if len(room.Players) == 3 {
					room.Status = "playing"
					// 发牌
					room.InitDeck()
					go room.StartGame()
				}
				gs.mutex.Unlock()
				hasRoom = true
				break
			}
		}
		if !hasRoom {
			// 创建新房间
			newRoom := &Room{
				ID:     generateRoomID(),
				Status: "waiting",
				LastUserId: 0,
				LastCards: make([]Card, 0),
				LastIsPass: false,
				CurSendCardUserId: 0,
			}
			player.RooId = newRoom.ID
			newRoom.Players = append(newRoom.Players, player)
			player.Conn.WriteJSON(player)
			gs.Rooms[newRoom.ID] = newRoom
			gs.mutex.Unlock()
		}
	}
}

func (r *Room) StartGame() {
	// d:=rand.Intn(2)
	for _, player := range r.Players {
		// 给每个玩家发18张牌
		hand := make([]Card, 18)
		for i := 0; i < 18; i++ {
			hand[i] = r.Deck[len(r.Deck)-1]
			r.Deck = r.Deck[:len(r.Deck)-1]
		}
		player.Conn.WriteJSON(map[string]interface{}{
			"type": "deal",
			"hand": hand,
		})
	}
}

func generateRoomID() string {
	return "ROOM-" + fmt.Sprintf("%d", time.Now().UnixNano())
}
// func (gs *GameServer) HandleMessage() {
// 	for {
// 		msg := <-gs.Msg
// 		roomID := msg.RoomId
// 		fmt.Println("处理打牌逻辑！")
// 		gs.mutex.RLock()
// 		if roomID != "" {
// 			room, exists := gs.Rooms[roomID]
// 			if exists {
// 				var curP *Player
// 				curP = nil
// 				room.mutex.Lock()
// 				for _, p := range room.Players {
// 					if p.ID == msg.UserId {
// 						curP = p
// 						break
// 					}
// 				}
// 				if curP == nil {
// 					fmt.Println("当前发送人ID不正确！")
// 				}
// 				room.mutex.Unlock()
// 			} else {
// 				fmt.Println("不存在该房间")

// 				//Output:

// 			}
// 		}

// 		gs.mutex.RUnlock()
// 	}
// }
func ResponseResult(conn  *websocket.Conn,res any){
	conn.WriteJSON(res)
}
func (gs *GameServer) HandleConnections(w http.ResponseWriter, r *http.Request) {
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Fatal(err)
	}
	defer ws.Close()
	userId := userNumber
	userNumber++
	player := &Player{Conn: ws, ID: userId, Name: fmt.Sprintf("玩家%d", userId)}
	gs.Lobby <- player
	for {
		// 合并变量声明和赋值
		msg := &Message{}
		err := ws.ReadJSON(msg)
		fmt.Printf("当前出牌玩家：%v,出的牌：%v\n", player,msg)
		if err != nil {
			log.Printf("error: %v", err)
			delete(gs.Rooms, player.RooId)
			break
		}
		responseMsg:=ResponseMessage{}
		curRoom:=gs.Rooms[player.RooId]
		if curRoom.LastUserId!=0&&player.ID!=curRoom.CurSendCardUserId{
			//出牌人不是该玩家
			responseMsg.SendCardStatus=-2
			responseMsg.Describe="不该你出牌！"
			ResponseResult(player.Conn,responseMsg)
			continue
		}
		//处理打牌逻辑
		//组装响应其他玩家的消息体
		sendCardLen:=len(msg.SendCard)
		receptionPlayers:=[]Player{}
		curPlayIndex:=0
		for index,roomPlay:=range curRoom.Players{
			if player.ID!=roomPlay.ID{
				receptionPlayers=append(receptionPlayers, *roomPlay)
			}else{
				curPlayIndex=index
			}
		}
		if curPlayIndex==2{
			responseMsg.NextUserId=curRoom.Players[0].ID
		}else{
			responseMsg.NextUserId=curRoom.Players[curPlayIndex+1].ID
		}
		curRoom.CurSendCardUserId=responseMsg.NextUserId
		//当前玩家 Pass
		if sendCardLen==0 {
			curRoom.LastIsPass=true
			responseMsg.SendCardStatus=0
			for _,otherPlay:=range receptionPlayers{
				ResponseResult(otherPlay.Conn,responseMsg)
			}
			continue;
		}
		sort.Sort(msg.SendCard)
		ranks := []string{"A", "2", "3", "4", "5", "6", "7", "8", "9", "I", "J", "Q", "K"}
		mapRanks:=make(map[string]int,0)
		for index,rank:=range ranks{
			mapRanks[rank]=index+1
		}
		//上一个出牌的id为空，则为第一次出牌
		if curRoom.LastUserId==0{
			curRoom.LastUserId=userId
			curRoom.LastCards=msg.SendCard
			curRoom.LastIsPass=false
		}else{
			//当前玩家出牌之后，其他玩家pass
			if curRoom.LastIsPass&&curRoom.LastUserId==userId{
				curRoom.LastUserId=userId
				curRoom.LastCards=msg.SendCard
				curRoom.LastIsPass=false
			}else{
				//开始比较当前出的牌和上一次出牌大小
				//首先比较牌的数量是否一致
				curLen:=len(msg.SendCard)
				//判断是否是王炸
				if curLen==2&&msg.SendCard[0].Value==14&&msg.SendCard[1].Value==15{
					curRoom.LastUserId=userId
					curRoom.LastCards=msg.SendCard
					curRoom.LastIsPass=false
				}else{
					//格式不对，重新出牌
					if curLen!=len(curRoom.LastCards){
						responseMsg.SendCardStatus=-1
						responseMsg.Describe="格式不对，请重新出牌！"
						ResponseResult(player.Conn,responseMsg)
						continue;
					}
					var lastStr,curStr string
					var lastValue,curValue int
					for _,curCard:=range msg.SendCard{
						curStr+=curCard.Rank
						curValue+=curCard.Value
					}
					for _,curCard:=range curRoom.LastCards{
						lastStr+=curCard.Rank
						lastValue+=curCard.Value
					}
					//判断是否是特殊情况
					if curLen>3&&curLen<7{
						var allChar string="A23456789IJQK";
						//匹配是否是四带一或者四带二
						re:=regexp.MustCompile(fmt.Sprintf("[%s]{4}",allChar))
						//匹配是否是三带一或者三带二
						re1:=regexp.MustCompile(fmt.Sprintf("[%s]{3}",allChar))
						reFindCurStr:=re.FindString(curStr)
						reFindLastStr:=re.FindString(lastStr)
						re1FindCurStr:=re1.FindString(curStr)
						re1FindLastStr:=re1.FindString(lastStr)
						if strings.EqualFold(reFindCurStr,reFindLastStr){
							curValue=mapRanks[string(reFindCurStr[0])]
							lastValue=mapRanks[string(reFindLastStr[0])]
						}else if strings.EqualFold(re1FindCurStr,re1FindLastStr){
							curValue=mapRanks[string(re1FindCurStr[0])]
							lastValue=mapRanks[string(re1FindLastStr[0])]
						}
					}
					//成功出牌
					if curValue>lastValue{
						//记录上一次出牌和上一次出牌人
						curRoom.LastUserId=userId
						curRoom.LastCards=msg.SendCard
						curRoom.LastIsPass=false
						responseMsg.SendCardStatus=2
					}else{
						//压不起，重新出牌
						responseMsg.SendCardStatus=1
						responseMsg.Describe="压不过上一次牌，请重新出牌"
						ResponseResult(player.Conn,responseMsg)
						continue;
					}
				}
			}
		}
		//组装结果消息，发送给其他两人
		responseMsg.CurSendCard=msg.SendCard
		responseMsg.SendCardStatus=2
		responseMsg.Describe="继续出牌！"
		for _,otherPlay:=range receptionPlayers{
			ResponseResult(otherPlay.Conn,responseMsg)
		}
		
		// gs.Msg <- msg
	}
}

func main() {
	gs := NewGameServer()
	http.HandleFunc("/ws", gs.HandleConnections)
	// go gs.HandleMessage()
	log.Println("http service started on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
