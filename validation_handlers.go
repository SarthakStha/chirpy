package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"
)

func handlerReadiness(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	if _, err := w.Write([]byte("OK\n")); err != nil {
		log.Printf("error writing response: %s", err)
	}
}

func handlerValidate(w http.ResponseWriter, r *http.Request) {
	type requestBody struct {
		Body string `json:"body"`
	}

	type responseBody struct {
		CleanedBody string `json:"cleaned_body"`
	}

	param := requestBody{}
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&param); err != nil {
		respondWithError(w, http.StatusBadRequest, "Could not decode request body", err)
		return
	}

	if len(param.Body) > 140 {
		respondWithError(w, http.StatusBadRequest, "Chirp is too long", nil)
		return
	}

	profaneWords := map[string]struct{}{
		"kerfuffle": {},
		"sharbert":  {},
		"fornax":    {},
	}

	words := strings.Split(param.Body, " ")
	for i, word := range words {
		if _, ok := profaneWords[strings.ToLower(word)]; ok {
			words[i] = "****"
		}
	}

	respondWithJSON(w, http.StatusOK, responseBody{CleanedBody: strings.Join(words, " ")})
}
