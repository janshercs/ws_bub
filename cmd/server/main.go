package main

import (
	"fmt"
	"log"
	"net/http"
	"server"
)

const dbFileName = "threads.db.json"

func main() {
	store, closeDB, err := server.NewFFSFromPath(dbFileName)
	if err != nil {
		log.Fatal(err)
	}
	defer closeDB()

	webserver := server.NewServer(store)
	handler := http.HandlerFunc(webserver.ServeHTTP)
	fmt.Printf("Starting server at http://localhost:5000\n")
	log.Fatal(http.ListenAndServe(":5000", handler))
}
