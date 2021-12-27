package server

import (
	"encoding/json"
	"fmt"
	"io"
)

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

func GetThreadsFromReader(rdr io.Reader) (Threads, error) {
	var d Threads
	err := json.NewDecoder(rdr).Decode(&d)
	if err != nil {
		err = fmt.Errorf("problem parsing thread, %v", err)
	}
	return d, err
}
