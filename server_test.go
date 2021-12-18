package server_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"reflect"
	"server"
	"strings"
	"testing"

	"github.com/gorilla/websocket"
)

func TestServer(t *testing.T) {
	t.Run("GET request to root works", func(t *testing.T) {
		request := newGETRequest("/")
		response := httptest.NewRecorder()

		store := &spyStore{}
		server := server.NewServer(store)

		server.ServeHTTP(response, request)
		assertStatus(t, response, http.StatusOK)
	})

	t.Run("Post 2 threads and GET request to /thread returns both threads", func(t *testing.T) {
		testServer := server.NewServer(&spyStore{})

		firstThreadPayload := newThreadPayload("this is thread 1", "anna")
		secondThreadPayload := newThreadPayload("this is thread 2", "bob")

		testcases := []struct {
			threadPayload threadPayload
		}{
			{firstThreadPayload},
			{secondThreadPayload},
		}

		for i, tc := range testcases {
			t.Run(fmt.Sprintf("post # %d", i), func(t *testing.T) {
				response := httptest.NewRecorder()
				request := newPOSTRequest("/thread", tc.threadPayload)

				testServer.ServeHTTP(response, request)

				assertStatus(t, response, http.StatusOK)
				assertThreadExceptID(t, getThreadFromBody(t, response.Body), threadPayloadToThread(tc.threadPayload))
			})
		}

		request := newGETRequest("/thread")
		response := httptest.NewRecorder()

		testServer.ServeHTTP(response, request)

		d := getThreadsFromBody(t, response.Body)

		assertThreads(t, d, []server.Thread{threadPayloadToThread(firstThreadPayload), threadPayloadToThread(secondThreadPayload)})

	})

	t.Run("Post empty thread and receive an error", func(t *testing.T) {
		testThread := newThreadPayload("", "anna")

		request := newPOSTRequest("/thread", testThread)
		response := httptest.NewRecorder()
		store := &spyStore{}
		testServer := server.NewServer(store)

		testServer.ServeHTTP(response, request)

		assertStatus(t, response, http.StatusBadRequest)
		assertError(t, response, server.EmptyContentErr)

		if len(store.threads) != 0 {
			t.Errorf("Should not have stored bad thread, but it did.")
		}
	})

	t.Run("Post thread with no user and receive an error", func(t *testing.T) {
		testThread := newThreadPayload("this is thread 1", "")

		request := newPOSTRequest("/thread", testThread)
		response := httptest.NewRecorder()
		store := &spyStore{}
		testServer := server.NewServer(store)

		testServer.ServeHTTP(response, request)

		assertStatus(t, response, http.StatusBadRequest)
		assertError(t, response, server.MissingUserErr)

		if len(store.threads) != 0 {
			t.Errorf("Should not have stored bad thread, but it did.")
		}
	})

	t.Run("GET request to /thread/0 returns first thread", func(t *testing.T) {
		thread := threadPayloadToThread(newThreadPayload("this is thread 1", "anna"))
		secondThread := threadPayloadToThread(newThreadPayload("this is thread 2", "karenina"))

		store := &spyStore{
			threads: []server.Thread{thread, secondThread},
		}
		testServer := server.NewServer(store)

		request := newGETRequest("/thread/0")
		response := httptest.NewRecorder()

		testServer.ServeHTTP(response, request)

		d := getThreadFromBody(t, response.Body)

		assertThreadExceptID(t, d, thread)

		request = newGETRequest("/thread/1")
		response = httptest.NewRecorder()

		testServer.ServeHTTP(response, request)

		d = getThreadFromBody(t, response.Body)

		assertThreadExceptID(t, d, secondThread)
	})

	t.Run("Invalid GET requests to /thread/{id} returns error", func(t *testing.T) {

		store := &spyStore{}
		testServer := server.NewServer(store)
		testcase := []struct {
			name string
			url  string
			err  error
			code int
		}{
			{
				name: "Getting from empty store",
				url:  "/thread/0",
				err:  server.MissingThreadErr,
				code: http.StatusNotFound,
			},
			{
				name: "Getting with string ID",
				url:  "/thread/a",
				err:  server.InvalidIDErr,
				code: http.StatusBadRequest,
			},
			{
				name: "Getting with negative ID",
				url:  "/thread/-1",
				err:  server.InvalidIDErr,
				code: http.StatusBadRequest,
			},
		}

		for _, tc := range testcase {
			t.Run(tc.name, func(t *testing.T) {
				request := newGETRequest(tc.url)
				response := httptest.NewRecorder()

				testServer.ServeHTTP(response, request)
				assertStatus(t, response, tc.code)
				assertError(t, response, tc.err)
			})
		}

	})
}

func TestWebSocket(t *testing.T) {
	threads := []server.Thread{
		{
			Content:        "What is the truth?",
			User:           "Neo",
			UpVotesCount:   0,
			DownVotesCount: 0,
		},
		{
			Content:        "There is no spoon",
			User:           "Random kid",
			UpVotesCount:   0,
			DownVotesCount: 0,
		},
	}

	testStore := &spyStore{threads}
	threadServer := server.NewServer(testStore)
	go threadServer.StartWorkers()

	testServer := httptest.NewServer(threadServer)
	wsURL := "ws" + strings.TrimPrefix(testServer.URL, "http") + "/ws"
	ws := MustDialWS(t, wsURL)

	defer ws.Close()
	defer testServer.Close()
	t.Run("Websocket request to /ws returns list all threads inside.", func(t *testing.T) {
		var got []server.Thread
		ws.ReadJSON(&got)
		assertThreads(t, got, threads)

	})

	t.Run("Websocket send Threads to /ws received and saved by store.", func(t *testing.T) {
		firstThreadPayload := newThreadPayload("Excited about the Matrix", "Trinity")
		threads = append(threads, threadPayloadToThread(firstThreadPayload))

		ws.WriteJSON(firstThreadPayload)

		var got []server.Thread
		ws.ReadJSON(&got)
		assertThreads(t, got, threads)

		secondThreadPayload := newThreadPayload("I know kungfu", "Neo")
		threads = append(threads, threadPayloadToThread(secondThreadPayload))
		ws.WriteJSON(secondThreadPayload)

		ws.ReadJSON(&got)
		assertThreads(t, got, threads)
	})

}

func newGETRequest(path string) *http.Request {
	request, _ := http.NewRequest(http.MethodGet, path, nil)
	return request
}

func newPOSTRequest(path string, data interface{}) *http.Request {
	payloadBuf := new(bytes.Buffer)
	err := json.NewEncoder(payloadBuf).Encode(data)
	if err != nil {
		log.Fatal(err)
	}

	request, _ := http.NewRequest(http.MethodPost, path, payloadBuf)
	request.Header.Set("Content-Type", server.JSONContentType)

	return request
}

func newThreadPayload(content, user string) threadPayload {
	return threadPayload{
		Content: content,
		User:    user,
	}
}

func assertBodyString(t testing.TB, response *httptest.ResponseRecorder, want string) {
	got := response.Body.String()
	if got != want {
		t.Errorf("wanted: %q, got %q", want, got)
	}
}

func assertError(t testing.TB, response *httptest.ResponseRecorder, err error) {
	got := strings.TrimSuffix(response.Body.String(), "\n")

	if got != err.Error() {
		t.Errorf("wrong error message received: got %q, but wanted %q", got, err.Error())
	}
}

func assertStatus(t testing.TB, got *httptest.ResponseRecorder, want int) {
	if got.Code != want {
		t.Errorf("status code is wrong: got status %d want %d", got.Code, want)
	}
}

func assertThreads(t testing.TB, got, want []server.Thread) {
	if len(got) != len(want) {
		t.Fatalf("Different number of threads returned, wanted %d, got %d.", len(want), len(got))
	}
	for i := 0; i < len(want); i++ {
		assertThreadExceptID(t, got[i], want[i])
	}
}

func assertThreadExceptID(t testing.TB, got, want server.Thread) {
	w := reflect.ValueOf(want)
	g := reflect.ValueOf(got)

	for i := 0; i < w.NumField(); i++ {
		fieldName := w.Type().Field(i).Name
		if fieldName == "ID" {
			continue
		}

		if g.Field(i).Interface() != w.Field(i).Interface() {
			t.Errorf("got %v want %v", got, want)
		}
	}
}

func getThreadFromBody(t testing.TB, r io.Reader) server.Thread {
	t.Helper()

	d, err := server.GetThreadFromReader(r)

	if err != nil {
		t.Errorf("Error occured while getting thread from response body: %v", err)
	}

	return d
}

func getThreadsFromBody(t testing.TB, r io.Reader) []server.Thread {
	t.Helper()
	var d []server.Thread

	err := json.NewDecoder(r).Decode(&d)
	if err != nil {
		t.Errorf("Error occured while getting thread from response body: %v", err)
	}

	return d
}

func threadPayloadToThread(tp threadPayload) server.Thread {
	return server.Thread{
		Content:        tp.Content,
		User:           tp.User,
		UpVotesCount:   0,
		DownVotesCount: 0,
	}
}

type threadPayload struct {
	Content string
	User    string
}

type spyStore struct {
	threads []server.Thread
}

func (s *spyStore) SaveThread(thread server.Thread) {
	s.threads = append(s.threads, thread)
}

func (s *spyStore) GetThreads() []server.Thread {
	return s.threads
}

func MustDialWS(t *testing.T, url string) *websocket.Conn {

	ws, _, err := websocket.DefaultDialer.Dial(url, http.Header{
		"Origin": []string{"http://localhost:3000"},
	})
	if err != nil {
		t.Fatalf("could not open a ws connection on %s %v", url, err)
	}
	return ws
}
