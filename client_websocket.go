package server

import "github.com/gorilla/websocket"

type ClientWS struct {
	socket *websocket.Conn
}

func (c ClientWS) SendThreads(t Threads) error {
	err := c.socket.WriteJSON(t)
	return err

}

func (c ClientWS) GetThread(t *Thread) error {
	err := c.socket.ReadJSON(t)
	return err
}
