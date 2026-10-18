package main

import (
	"log"
	"net/http"
)

func handlerHealthz(resp http.ResponseWriter, req *http.Request) {
	resp.Header().Set("Content-Type", "text/plain; charset=utf-8")
	resp.WriteHeader(200)
	if _, err := resp.Write([]byte("OK")); err != nil {
		log.Fatal(err)
	}
}
