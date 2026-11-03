package main

import (
	"encoding/json"
	"github.com/google/uuid"
	"net/http"
	"time"
)

func (cfg *apiConfig) handlerNewUser(w http.ResponseWriter, r *http.Request) {
	userEmail := struct {
		Email string `json:"email"`
	}{}

	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&userEmail); err != nil {
		respondWithError(w, http.StatusBadRequest, "Could not decode request body", err)
		return
	}

	user, err := cfg.db.CreateUser(r.Context(), userEmail.Email)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Unable to create user", err)
		return
	}

	// This is sending the direct content of User
	// Might need to be wrapped in a user struct Eg: {user: {id:   , created_at:....}}
	responseJSON := struct {
		ID        uuid.UUID `json:"id"`
		CreatedAt time.Time `json:"created_at"`
		UpdatedAt time.Time `json:"updated_at"`
		Email     string    `json:"email"`
	}{
		ID:        user.ID,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
		Email:     user.Email,
	}

	respondWithJSON(w, http.StatusCreated, responseJSON)
}
