package main

import (
	"fmt"
	"log"
	"net/http"
	"server"
)

func main() {
	webserver := server.NewServer()
	handler := http.HandlerFunc(webserver.ServeHTTP)
	fmt.Printf("Starting server at http://localhost:5000\n")
	log.Fatal(http.ListenAndServe(":5000", handler))
}
