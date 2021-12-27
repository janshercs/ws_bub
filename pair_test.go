package server_test

import (
	"net/http/httptest"
	"reflect"
	"server"
	"strings"
	"testing"

	"github.com/gorilla/websocket"
)

func TestPairWS(t *testing.T) {
	testStore := &spyStore{}

	threadServer := server.NewServer(testStore, NewSpyClientManager())
	go threadServer.StartWorkers()

	testServer := httptest.NewServer(threadServer)
	wsURL := "ws" + strings.TrimPrefix(testServer.URL, "http") + "/pair"
	ws := MustDialWS(t, wsURL)

	defer ws.Close()
	defer testServer.Close()
	t.Run("Receives default value", func(t *testing.T) {
		welcomeMessage := []byte("hi, enter text here")
		_, msg, err := ws.ReadMessage()

		if err != nil {
			t.Errorf("Error reading message, %v", err)
		}

		assertRightMessage(t, welcomeMessage, msg)
	})

	t.Run("Changes value with new messages sent", func(t *testing.T) {
		firstMessage := []byte("this is a sample message")
		ws.WriteMessage(websocket.TextMessage, firstMessage)

		_, msg, _ := ws.ReadMessage()
		assertRightMessage(t, firstMessage, msg)

		secondMessage := append(firstMessage, []byte("/n another message behind.")...)
		ws.WriteMessage(websocket.TextMessage, secondMessage)

		_, msg, _ = ws.ReadMessage()
		assertRightMessage(t, secondMessage, msg)
	})

}

func assertRightMessage(t *testing.T, want, got []byte) {
	if !reflect.DeepEqual(got, want) {
		t.Errorf("Did not receieve the right message, wanted %s, got %s", want, got)
	}
}
