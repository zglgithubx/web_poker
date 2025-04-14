package data_struct

import (
	"github.com/gorilla/websocket"
	ut "mq/utils"
	"sync"
)

type GameServer struct {
	Rooms map[string]*Room
	Lobby chan *Player
	Msg   chan *TransferMessage
	Mutex sync.RWMutex
}
type Room struct {
	ID                string
	Players           map[string]*Player
	PlayersIndex      map[string]int
	Deck              []Card
	Mutex             sync.Mutex
	Status            string // waiting/playing
	LastCards         Cards
	LastUserId        string
	LastIsPass        bool
	CurSendCardUserId string
}

// 生成牌
func (r *Room) InitDeck() {
	r.Mutex.Lock()
	defer r.Mutex.Unlock()
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
func (r *Room) StartGame() {
	//发牌
	for _, player := range r.Players {
		// 给每个玩家发18张牌
		hand := make([]Card, 18)
		for i := 0; i < 18; i++ {
			hand[i] = r.Deck[len(r.Deck)-1]
			r.Deck = r.Deck[:len(r.Deck)-1]
		}
		data := make(map[string]any, 2)
		data["playerHand"] = hand
		data["dealerHand"] = hand
		ut.ResponseResult(player.Conn, map[string]interface{}{
			"type": "gameStart",
			"data": data,
		})
	}
}

type Player struct {
	Conn   *websocket.Conn
	RooId  string
	ID     string
	Name   string
	UserId string
}
type Card struct {
	Suit  string `json:"suit"`
	Rank  string `json:"rank"`
	Value int    `json:"value"`
}
type Cards []Card

func (c Cards) Len() int {
	return len(c)
}
func (c Cards) Less(i, j int) bool {
	return c[i].Value < c[j].Value
}
func (c Cards) Swap(i, j int) {
	c[i], c[j] = c[j], c[i]
}
