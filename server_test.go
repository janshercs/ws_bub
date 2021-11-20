package server_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"server"
	"testing"
)

func TestServer(t *testing.T) {
	// t.Run("GET request to root works", func(t *testing.T) {
	// 	request := newGETRequest("/")
	// 	response := httptest.NewRecorder()

	// 	store := &spyStore{}
	// 	server := server.NewServer(store)

	// 	server.ServeHTTP(response, request)
	// 	assertStatus(t, response, http.StatusOK)
	// })

	t.Run("Post thread and GET request to /thread returns thread content", func(t *testing.T) {
		want := "this is thread 1"
		postData := struct {
			Thread string
			User   string
		}{want, "jansen"}

		payloadBuf := new(bytes.Buffer)

		json.NewEncoder(payloadBuf).Encode(postData)

		request, _ := http.NewRequest(http.MethodPost, "/thread", payloadBuf)
		request.Header.Set("Content-Type", "application/json")
		response := httptest.NewRecorder()

		store := &spyStore{}
		testServer := server.NewServer(store)

		testServer.ServeHTTP(response, request)
		assertStatus(t, response, http.StatusOK)

		request = newGETRequest("/thread")
		response = httptest.NewRecorder()

		testServer.ServeHTTP(response, request)
		assertBodyString(t, response, want)
	})
}

func newGETRequest(path string) *http.Request {
	request, _ := http.NewRequest(http.MethodGet, path, nil)
	return request
}

func assertStatus(t testing.TB, got *httptest.ResponseRecorder, want int) {
	if got.Code != want {
		t.Errorf("status code is wrong: got status %d want %d", got.Code, want)
	}
}

func assertBodyString(t testing.TB, response *httptest.ResponseRecorder, want string) {
	got := response.Body.String()
	if got != want {
		t.Errorf("wanted: %q, got %q", want, got)
	}
}

type spyStore struct {
	threads []string
}

func (s *spyStore) SaveThread(thread string) {
	s.threads = append(s.threads, thread)
}

func (s *spyStore) GetThreads() []string {
	return s.threads
}
