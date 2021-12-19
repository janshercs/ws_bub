package server

import (
	"log"

	"github.com/gorilla/websocket"
)

type ClientWS struct {
	socket *websocket.Conn
}

func (c ClientWS) SendThreads(t Threads) error {
	err := c.socket.WriteJSON(t)
	return err

}

func (c ClientWS) GetThread() (Thread, error) {
	var t Thread
	err := c.socket.ReadJSON(&t)
	if err != nil {
		return t, err

	}

	return t, nil
}

type WebSocketManager interface {
	RemoveClient(client *ClientWS)
	AddClient(client *ClientWS)
	Broadcast(Threads)
}

type ClientManager struct {
	Clients map[*ClientWS]bool
}

func (c ClientManager) AddClient(client *ClientWS) {
	c.Clients[client] = true
	log.Printf("Added client %v.", client)

}

func (c ClientManager) RemoveClient(client *ClientWS) {
	delete(c.Clients, client)
	log.Printf("Removed client %v.", client)
}

func (c ClientManager) Broadcast(t Threads) {
	for client := range c.Clients {
		err := client.SendThreads(t)
		if err != nil {
			log.Printf("Error encountered when sending to client. %v", err) // TODO: Add missing client handling here.
		}
	}
}

func NewClientManager() *ClientManager {
	return &ClientManager{Clients: make(map[*ClientWS]bool)}
}
