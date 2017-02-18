package main

import (
	"time"
	"github.com/gorilla/websocket"
)

type client struct {
	// このクライアント用のWebSocket
	socket *websocket.Conn
	// メッセージが送られるチャネル
	send chan *message
	// クライアントが参加しているチャットルーム
	room *room
	// ユーザー情報
	userData map[string]interface{}
}

func (c *client) read()  {
	for {
		var msg *message
		if err := c.socket.ReadJSON(&msg); err == nil {
			msg.When = time.Now()
			msg.Name = c.userData["name"].(string)
			c.room.forward <- msg
		} else {
			break
		}
	}
	c.socket.Close()
}

func (c *client) write() {
	for msg := range c.send {
		if err := c.socket.WriteJSON(msg); err != nil {
			break
		}
	}
	c.socket.Close()
}
