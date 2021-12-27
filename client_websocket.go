package server

import (
	"log"

	"github.com/gorilla/websocket"
)

type ClientWS struct {
	socket *websocket.Conn
	pair   bool
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

func (c ClientWS) WriteMessage(msg []byte) error {
	err := c.socket.WriteMessage(websocket.TextMessage, msg)
	return err
}

type WebSocketManager interface {
	RemoveClient(client *ClientWS)
	AddClient(client *ClientWS)
	Broadcast([]*ClientWS, interface{})
	GetPairClients() []*ClientWS
	GetChatClients() []*ClientWS
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

func (c ClientManager) Broadcast(clients []*ClientWS, payload interface{}) {
	switch payload.(type) {
	case Threads:
		for _, client := range clients {
			threads := payload.(Threads)
			err := client.SendThreads(threads)
			if err != nil {
				log.Printf("Error encountered when sending to client. %v", err) // TODO: Add missing client handling here.
			}
		}

	case []byte:
		for _, client := range clients {
			msg := payload.([]byte)
			err := client.WriteMessage(msg)
			if err != nil {
				log.Printf("Error encountered when sending to client. %v", err)
			}
		}
	}
}

func (c ClientManager) GetPairClients() []*ClientWS {
	var clients []*ClientWS
	for client := range c.Clients {
		if !client.pair {
			continue
		}
		clients = append(clients, client)
	}
	return clients
}

func (c ClientManager) GetChatClients() []*ClientWS { // FIXME: Need to change naming!!
	var clients []*ClientWS
	for client := range c.Clients {
		if client.pair {
			continue
		}
		clients = append(clients, client)
	}
	return clients
}

func NewClientManager() *ClientManager {
	return &ClientManager{Clients: make(map[*ClientWS]bool)}
}
