package server

import "github.com/gorilla/websocket"

type ClientWS struct {
	socket *websocket.Conn
}

func (c ClientWS) SendThreads(t Threads) {
	c.socket.WriteJSON(t)
}

func (c ClientWS) GetThread(t *Thread) {
	c.socket.ReadJSON(t)
}
