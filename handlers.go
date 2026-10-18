package main

import (
	"log"
	"net/http"
)

func handlerReadiness(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(200)
	if _, err := w.Write([]byte("OK\n")); err != nil {
		log.Printf("error writing response: %s", err)
	}
}
