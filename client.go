package main

import "github.com/gorilla/websocket"

// a client is a single chatting user !-! on server side !-!
type client struct {
	socket *websocket.Conn

	// 要发送的信息
	send chan []byte

	// 注意这里，同是main包的，不需要引入就直接用了！
	room *room
}

func (c *client) read()  {
	defer c.socket.Close()
	for {
		_, msg, err := c.socket.ReadMessage()
		if err != nil {
			return
		}
		c.room.forward <- msg
	}
}

func (c *client) write()  {
	defer c.socket.Close()
	for msg := range c.send {
		err := c.socket.WriteMessage(websocket.TextMessage, msg)
		if err != nil{
			return
		}
	}
}