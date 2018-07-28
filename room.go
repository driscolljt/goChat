package main

import (
	"log"
	"net/http"

	"github.com/stretchr/objx"

	"github.com/gorilla/websocket"

)

type room struct {
	// forward is a channel that holds incoming messages
	// that should be forwarded to the other clients
	forward chan *message
	// join is a channel for clients wishing to join the room
	join chan *client
	// leave is a channel for clients wishing to leave the room
	leave chan *client
	// clients holds all current clients in this room
	clients map[*client]bool
}

// newRoom makes a new room
func newRoom() *room {
	return &room{
		forward: make(chan *message),

		join: make(chan *client),

		leave: make(chan *client),

		clients: make(map[*client]bool),
	}
}

func (r *room) run() {
	for {
		select {
		case client := <-r.join:
			// joining
			r.clients[client] = true
			log.Println("New client joined!")
		case client := <-r.leave:
			// leaving
			delete(r.clients, client)
			close(client.send)
			log.Println("Client left")
		case msg := <-r.forward:
			log.Println("Message received: '", string(msg.Message), "' at: '", msg.When, "'" )
			// forward message to all clients
			for client := range r.clients {
				client.send <- msg
				log.Println(" -- sent to client: ", client.userData["name"])
			}
		}
	}
}

const (
	socketBufferSize  = 1024
	messageBufferSize = 256
)

var upgrader = &websocket.Upgrader{ReadBufferSize: socketBufferSize, WriteBufferSize: socketBufferSize}

func (r *room) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	socket, err := upgrader.Upgrade(w, req, nil)
	if err != nil {
		log.Fatal("ServerHTTP:", err)
		return
	}
	authCookie, err := req.Cookie("auth")
	if err != nil {
		log.Fatal("Failed to get auth cookie:", err)
		return
	}
	client := &client{
		socket:   socket,
		send:     make(chan *message, messageBufferSize),
		room:     r,
		userData: objx.MustFromBase64(authCookie.Value),
	}
	r.join <- client
	defer func() { r.leave <- client }()
	go client.write()
	client.read()
}
