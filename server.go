package server

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"path"
	"strconv"
)

const (
	JSONContentType         = "application/json"
	UnreadablePayloadErrMsg = "Unable to decode payload"
)

var (
	InvalidIDErr     = errors.New("Invalid ID provided")
	EmptyContentErr  = errors.New("Thread content must have at least 1 character.")
	MissingUserErr   = errors.New("Thread is missing a user.")
	MissingThreadErr = errors.New("The thread you are looking for does not exists.")
)

type ThreadStore interface {
	SaveThread(thread Thread)
	GetThreads() []Thread
}

type Server struct {
	store ThreadStore
	http.Handler
}

func NewServer(store ThreadStore) *Server {
	s := new(Server)
	s.store = store

	router := http.NewServeMux()
	router.Handle("/", http.HandlerFunc(s.homeHandler))
	router.Handle("/thread", http.HandlerFunc(s.threadHandler))
	router.Handle("/thread/", http.HandlerFunc(s.singleThreadHandler))

	s.Handler = router

	return s
}

func (s *Server) homeHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hello!")
}

func (s *Server) threadHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {

	case http.MethodPost:
		thread, err := GetThreadFromBody(r.Body)
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
	json.NewEncoder(w).Encode(s.store.GetThreads()[index])

}

type Thread struct {
	ID             int
	Content        string
	User           string
	UpVotesCount   int
	DownVotesCount int
}

func GetThreadFromBody(rdr io.Reader) (Thread, error) {
	var d Thread
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
