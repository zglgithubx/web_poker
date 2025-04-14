package utils

import "github.com/gorilla/websocket"

func ResponseResult(conn *websocket.Conn, res any) {
	err := conn.WriteJSON(res)
	if err != nil {
		return
	}
}
