package server

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"path"
	"strconv"

	"github.com/gorilla/websocket"
)

const (
	JSONContentType         = "application/json"
	UnreadablePayloadErrMsg = "Unable to decode payload"
)

var (
	allowedOrigins   = []string{"http://localhost:3000", "https://wassup-bub.netlify.app"}
	InvalidIDErr     = errors.New("Invalid ID provided")
	EmptyContentErr  = errors.New("Thread content must have at least 1 character.")
	MissingUserErr   = errors.New("Thread is missing a user.")
	MissingThreadErr = errors.New("The thread you are looking for does not exists.")
	wsUpgrader       = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin: func(r *http.Request) bool {
			if OriginIsAllowed(r) {
				return true
			}
			fmt.Printf("Refused websocket connection to %v", r.Header.Get("Origin"))
			return false
		},
	}
)

type ThreadStore interface {
	SaveThread(thread Thread)
	GetThreads() []Thread
}

type Server struct {
	store ThreadStore
	http.Handler
	sockets []*ClientWS
}

func NewServer(store ThreadStore) *Server {
	s := new(Server)
	s.store = store

	router := http.NewServeMux()
	router.Handle("/", http.HandlerFunc(s.homeHandler))
	router.Handle("/thread", http.HandlerFunc(s.threadHandler))
	router.Handle("/thread/", http.HandlerFunc(s.singleThreadHandler))
	router.Handle("/ws", http.HandlerFunc(s.websocketHandler))

	s.Handler = router

	return s
}

func (s *Server) homeHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hello!")
}

func (s *Server) threadHandler(w http.ResponseWriter, r *http.Request) {
	if OriginIsAllowed(r) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Header().Set("Access-Control-Allow-Origin", r.Header.Get("Origin"))
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type,access-control-allow-origin, access-control-allow-headers")
	}
	switch r.Method {

	case http.MethodPost:
		thread, err := GetThreadFromReader(r.Body)
		if err != nil {
			http.Error(w, UnreadablePayloadErrMsg, http.StatusBadRequest)
		}

		err = s.checkThread(thread)

		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		threadID := len(s.store.GetThreads())
		thread.ID = threadID
		s.store.SaveThread(thread)

		json.NewEncoder(w).Encode(thread)

		for _, conn := range s.sockets {
			conn.SendThreads(s.store.GetThreads())
		}

	default:
		w.Header().Set("content-type", JSONContentType)
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(s.store.GetThreads())
	}
}

func (s *Server) singleThreadHandler(w http.ResponseWriter, r *http.Request) {
	index, err := s.GetIDFromRequest(r)

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	threads := s.store.GetThreads()

	if index >= len(threads) {
		http.Error(w, "The thread you are looking for does not exists.", http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(threads[index]) // TODO: add getThreadByID()

}

func (s *Server) websocketHandler(w http.ResponseWriter, r *http.Request) {
	client := NewClientWS(w, r)

	client.SendThreads(s.store.GetThreads())
	s.sockets = append(s.sockets, client)

	var t Thread
	go func() {
		for {
			client.GetThread(&t)  // TODO: feels like should create a channel where it will return thread
			s.store.SaveThread(t) // TODO: feels like should dequeue from the channel and store the thread
			for _, socket := range s.sockets {
				socket.SendThreads(s.store.GetThreads())
			}
		}
	}()
}

type Threads []Thread
type Thread struct {
	ID             int
	Content        string
	User           string
	UpVotesCount   int
	DownVotesCount int
}

func GetThreadFromReader(rdr io.Reader) (Thread, error) {
	var d Thread
	err := json.NewDecoder(rdr).Decode(&d)
	if err != nil {
		err = fmt.Errorf("problem parsing thread, %v", err)
	}
	return d, err
}

func GetThreadsFromReader(rdr io.Reader) ([]Thread, error) {
	var d []Thread
	err := json.NewDecoder(rdr).Decode(&d)
	if err != nil {
		err = fmt.Errorf("problem parsing thread, %v", err)
	}
	return d, err
}

func (s *Server) checkThread(thread Thread) error {
	if len(thread.Content) == 0 {
		return EmptyContentErr
	}

	if len(thread.User) == 0 {
		return MissingUserErr

	}

	return nil
}

func (s *Server) GetIDFromRequest(r *http.Request) (int, error) {
	index, err := strconv.Atoi(path.Base(r.URL.Path))

	if err != nil || index < 0 {
		return 0, InvalidIDErr
	}

	return index, nil
}

func OriginIsAllowed(r *http.Request) bool {
	requestOrigin := r.Header.Get("Origin")
	for _, origin := range allowedOrigins {
		if requestOrigin == origin {
			return true
		}
	}
	return false
}

func NewClientWS(w http.ResponseWriter, r *http.Request) *ClientWS {
	conn, err := wsUpgrader.Upgrade(w, r, nil)

	if err != nil {
		log.Printf("problem upgrading connection to Websockets %v\n", err)
	}
	return &ClientWS{conn}
}
