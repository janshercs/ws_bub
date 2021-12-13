package main

import (
	"log"
	"net/http"
	"os"
	"server"
)

const dbFileName = "threads.db.json"

func main() {

	port := os.Getenv("PORT")
	if port == "" {
		port = "5000"
	}
	store, closeDB, err := server.NewFFSFromPath(dbFileName)
	if err != nil {
		log.Fatal(err)
	}
	defer closeDB()

	webserver := server.NewServer(store)
	handler := http.HandlerFunc(webserver.ServeHTTP)
	log.Printf("Starting server at http://localhost:%s\n", port)
	log.Fatal(http.ListenAndServe(":"+port, handler))
}
