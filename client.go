package main

import (
	"time"

	"github.com/gorilla/websocket"
)

// client represents a single chatting user
type client struct {
	// socket is the web socket for this client
	socket *websocket.Conn

	// send is a channel on which messages are sent
	send chan *message

	// room is the room where this client is chatting in
	room *room

	// userDate holds information about the user
	userData map[string]interface{}
}

func (c *client) read() {
	defer c.socket.Close()
	for {
		var msg *message
		err := c.socket.ReadJSON(&msg)
		if err != nil {
			return
		}
		// msg.When = time.Now().Format("1/2/2006 3:04 PM") // format date ' M/d/yyyy H:mm PM '
		msg.When = time.Now().Format("3:04 PM") // format date ' H:mm PM '
		msg.Name = c.userData["name"].(string)
		c.room.forward <- msg
	}
}

func (c *client) write() {
	defer c.socket.Close()
	for msg := range c.send {
		err := c.socket.WriteJSON(msg)
		if err != nil {
			return
		}
	}
}
