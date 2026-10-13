package main

import (
	"fmt"
	"log"
	"net/http"
	"time"
)

func main() {
	const port = "8080"
	// Currently has no routes
	mux := http.NewServeMux()

	server := &http.Server{
		Addr:           ":" + port,
		Handler:        mux,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	fmt.Println("Server is now running. Listening to port:", port)
	log.Fatal(server.ListenAndServe())
}
