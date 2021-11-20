package server

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type ThreadStore interface {
	SaveThread(thread string)
	GetThreads() []string
}

type Server struct {
	store ThreadStore
}

func NewServer(store ThreadStore) *Server {
	return &Server{store}
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch {

	case r.URL.Path == "/thread": // TODO: good time to split into handlers
		switch r.Method {

		case http.MethodPost:
			var d postData
			json.NewDecoder(r.Body).Decode(&d)
			s.store.SaveThread(d.Thread)

		default:
			for _, thread := range s.store.GetThreads() {
				fmt.Fprintf(w, thread) // TODO: change to json reply in the future.
			}
		}

	default:
		fmt.Fprintf(w, "Hello!")
	}
}

type postData struct {
	Thread string
	User   string
}
