package main

import (
	"net/http"
	"log"
	"github.com/gorilla/websocket"
)

const (
	socketBufferSize = 1024
	messageBufferSize = 256
)

type room struct {
	// forward is a channel that holds incoming messages
	// that should be forwarded to the other clients.
	forward chan []byte
	// join is a channel for clients wishing to join the room.
	join chan *client
	// leave is a channel for clients wishing to leave the room.
	leave chan *client
	// clients holds all current clients in this room.
	clients map[*client]bool
}

func newRoom()  *room{        // 有时，创建New方法，并不是因为需要检验值或者什么的，如下，只是为了少写几行初始化代码
	return &room{
		forward: make(chan []byte),
		join: make(chan *client),
		leave: make(chan *client),
		clients: make(map[*client]bool),
	}
}

//room会监测每个client的如下三个动作：加入，退出，发送消息
func (r *room) run() {
	for {
		select {
		case client := <-r.join:
			r.clients[client] = true
		case client := <-r.leave:
			delete(r.clients, client)
			close(client.send)
		case msg := <-r.forward:
			for client := range r.clients {
				client.send <- msg
			}
		}
	}
}

var upgrader = &websocket.Upgrader{
	ReadBufferSize: socketBufferSize,
	WriteBufferSize: socketBufferSize,
	// 若没有下面，报错 -> ServeHTTP:websocket: 'Origin' header value not allowed
	// https://godoc.org/github.com/gorilla/websocket#Upgrader
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func (r *room) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	// A server application uses the Upgrade function from an Upgrader object with a HTTP request handler
	// to get a pointer to a Conn:
	socket, err := upgrader.Upgrade(w, req, nil)
	if err != nil {
		log.Fatal("ServeHTTP:", err)
		return
	}
	client := &client{
		socket: socket,
		send: make(chan []byte, messageBufferSize),
		room: r,
	}
	r.join <- client
	defer func() { r.leave <- client }()
	go client.write()
	client.read() 	 //block operations (keeping the connection alive) until it's time to close it.
}