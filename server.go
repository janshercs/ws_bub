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
	http.Handler
	socketManager WebSocketManager
	store         ThreadStore
	threadChannel chan Thread
	sendChannel   chan bool
}

func NewServer(store ThreadStore, WSManager WebSocketManager) *Server {
	s := new(Server)

	s.store = store
	s.socketManager = WSManager
	s.threadChannel = make(chan Thread, 3)
	s.sendChannel = make(chan bool, 3)

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

		s.socketManager.Broadcast(s.store.GetThreads())

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
	s.socketManager.AddClient(client)
	log.Printf("Added socket %v.", client)

	go s.ProcessThreadFromClient(client)
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

func (s *Server) ProcessThreadFromClient(client *ClientWS) {
	for {
		t, err := client.GetThread()
		if err != nil {
			log.Println("Websocket closed.")
			s.socketManager.RemoveClient(client)
			return
		}

		threadErr := s.checkThread(t)
		if threadErr != nil {
			log.Println(err)
			return
		}
		s.threadChannel <- t
	}
}

func (s *Server) StartWorkers() {
	go s.ThreadSaver()
	go s.SocketUpdater()
}

func (s *Server) ThreadSaver() {
	for {
		t := <-s.threadChannel
		s.store.SaveThread(t)
		s.sendChannel <- true
	}
}

func (s *Server) SocketUpdater() {
	for {
		signal := <-s.sendChannel
		if signal {
			s.socketManager.Broadcast(s.store.GetThreads())
		}
	}
}
