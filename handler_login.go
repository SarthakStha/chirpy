package main

import (
	"encoding/json"
	"github.com/SarthakStha/chirpy/internal/auth"
	"github.com/google/uuid"
	"net/http"
	"time"
)

func (cfg *apiConfig) handlerLogin(w http.ResponseWriter, r *http.Request) {
	type params struct {
		Password  string        `json:"password"`
		Email     string        `json:"email"`
		ExpiresIn time.Duration `json:"expires_in_seconds"`
	}

	requestParams := params{}
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&requestParams); err != nil {
		respondWithError(w, http.StatusBadRequest, "Unable to decode request body", err)
		return
	}

	user, err := cfg.db.GetUserByEmail(r.Context(), requestParams.Email)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Unable to find a user with this credential", err)
		return
	}

	valid, err := auth.CheckPasswordHash(requestParams.Password, user.Password)
	if err != nil || !valid {
		respondWithError(w, http.StatusUnauthorized, "Unable to find a user with this credential", err)
		return
	}

	// If no expiration duration is provided use default 1 hour
	expirationTime := time.Hour * 1
	if requestParams.ExpiresIn != 0 {
		expirationTime = requestParams.ExpiresIn * time.Second
	}
	tokenString, err := auth.MakeJWT(user.ID, cfg.secret, expirationTime)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Unable to create a JWT for the user", err)
	}

	responseJSON := struct {
		ID        uuid.UUID `json:"id"`
		CreatedAt time.Time `json:"created_at"`
		UpdatedAt time.Time `json:"updated_at"`
		Email     string    `json:"email"`
		Token     string    `json:"token"`
	}{
		ID:        user.ID,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
		Email:     user.Email,
		Token:     tokenString,
	}
	respondWithJSON(w, http.StatusOK, responseJSON)
}
