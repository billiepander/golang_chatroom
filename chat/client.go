package main

import (
	"github.com/gorilla/websocket"
	"time"
)

// a client is a single chatting user !-! on server side !-!
type client struct {
	socket *websocket.Conn

	// 要发送的信息
	send chan *message

	// 注意这里，同是main包的，不需要引入就直接用了！
	room *room

	userData map[string]interface{}
}

func (c *client) read()  {
	defer c.socket.Close()
	for {
		var msg *message
		err := c.socket.ReadJSON(&msg)
		if err != nil {
			return
		}
		msg.When = time.Now()
		msg.Name = c.userData["name"].(string)
		if avatarURL, ok := c.userData["avatar_url"]; ok {
			msg.AvatarURL = avatarURL.(string)
		}
		c.room.forward <- msg
	}
}

func (c *client) write()  {
	defer c.socket.Close()
	for msg := range c.send {
		err := c.socket.WriteJSON(msg)
		if err != nil{
			return
		}
	}
}