package main

import (
	"encoding/json"
	"log"
	"net/http"
)

func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	w.Header().Set("Contect-Type", "application/json")

	resp, err := json.Marshal(payload)
	if err != nil {
		log.Printf("error marshalling JSON: %s", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(code)
	w.Write(resp)
}

func respondWithError(w http.ResponseWriter, code int, msg string, err error) {
	if err != nil {
		log.Printf("error while responding: %s", err)
	}

	if code > 499 {
		log.Printf("responding with 5XX error: %s", msg)
	}

	type errorResponse struct {
		Error string `json:"Error"`
	}

	respondWithJSON(w, code, errorResponse{Error: msg})
}
