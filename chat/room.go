package main

import (
	"net/http"
	"log"
	"github.com/gorilla/websocket"
	"github.com/chatroom/trace"
	"github.com/stretchr/objx"
)

const (
	socketBufferSize = 1024
	messageBufferSize = 256
)

type room struct {
	// forward is a channel that holds incoming messages
	// that should be forwarded to the other clients.
	forward chan *message
	// join is a channel for clients wishing to join the room.
	join chan *client
	// leave is a channel for clients wishing to leave the room.
	leave chan *client
	// clients holds all current clients in this room.
	clients map[*client]bool
	// tracer will receive trace information of activity in the room.
	tracer trace.Tracer
}

func newRoom()  *room{        // 有时，创建New方法，并不是因为需要检验值或者什么的，如下，只是为了少写几行初始化代码
	return &room{
		forward: make(chan *message),
		join: make(chan *client),
		leave: make(chan *client),
		clients: make(map[*client]bool),
		tracer: trace.Off(),     // 默认返回一个不记录的trace，这里是在main中改变为了要记录的情况; 当然也可以默认设置为trace.New(os.Stdout)，这样就是默认开启了
	}
}

//room会监测每个client的如下三个动作：加入，退出，发送消息
func (r *room) run() {
	for {
		select {
		case client := <-r.join:
			r.clients[client] = true
			r.tracer.Trace("New client joined")
		case client := <-r.leave:
			delete(r.clients, client)
			close(client.send)
			r.tracer.Trace("Client left")
		case msg := <-r.forward:
			r.tracer.Trace("Message received: ", msg.Message)
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
	authCookie, err := req.Cookie("auth")
	if err != nil {
		log.Fatal("Failed to get auth cookie:", err)
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
	client.read() 	 //block operations (keeping the connection alive) until it's time to close it.
}