package main

import (
	"net/http"
	"log"
	"github.com/stretchr/objx"
	"github.com/gorilla/websocket"
	"github.com/umatoma/chat/trace"
)

type room struct {
	// 他のクライアントに転送するためのメッセージを保持するチャネル
	forward chan *message
	// チャットルームに参加しようとしているクライアント用のチャネル
	join chan *client
	// チャットルームから退室しようとしているクライアント用のチャネル
	leave chan *client
	// 在室している全てのクライアントを保持
	clients map[*client]bool
	// チャットルーム上で行われた操作のログを受け取る
	tracer trace.Tracer
	// アバターの情報を取得します
	avatar Avatar
}

func newRoom(avatar Avatar) *room {
	return &room{
		forward: make(chan *message),
		join: make(chan *client),
		leave: make(chan *client),
		clients: make(map[*client]bool),
		tracer: trace.Off(),
		avatar: avatar,
	}
}

func (r *room) run() {
	for {
		select {
		case client := <- r.join:
			r.clients[client] = true
			r.tracer.Trace("新しいクライアントが参加しました")
		case client := <- r.leave:
			delete(r.clients, client)
			close(client.send)
			r.tracer.Trace("クライアントが退室しました")
		case msg := <- r.forward:
			r.tracer.Trace("メッセージを受信しました: ", msg.Message)
			// 全てのクライアントにメッセージを転送
			for client := range r.clients {
				select {
				case client.send <- msg:
					// メッセージを送信
					r.tracer.Trace(" -- クライアントに送信されました")
				default:
					// 送信に失敗
					delete(r.clients, client)
					close(client.send)
					r.tracer.Trace(" -- 送信に失敗しました。クライアントをクリーンアップします")
				}
			}
		}
	}
}

const (
	socketBufferSize = 1024
	messageBufferSize = 256
)
var upgrader = &websocket.Upgrader{
	ReadBufferSize: socketBufferSize,
	WriteBufferSize: messageBufferSize,
}
func (r *room) ServeHTTP(w http.ResponseWriter, req *http.Request)  {
	socket, err := upgrader.Upgrade(w, req, nil)
	if err != nil {
		log.Fatal("ServeHTTP:", err)
		return
	}

	authCookie, err := req.Cookie("auth")
	if err != nil {
		log.Fatal("クッキーの取得に失敗しました:", err)
		return
	}

	client := &client{
		socket: socket,
		send: make(chan *message, messageBufferSize),
		room: r,
		userData: objx.MustFromBase64(authCookie.Value),
	}
	r.join <- client
	defer func() { r.leave <- client }()
	go client.write()
	client.read()
}
