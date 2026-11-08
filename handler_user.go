package main

import (
	"encoding/json"
	"github.com/SarthakStha/chirpy/internal/auth"
	"github.com/SarthakStha/chirpy/internal/database"
	"github.com/google/uuid"
	"net/http"
	"time"
)

type User struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Email     string    `json:"email"`
}

func (cfg *apiConfig) handlerNewUser(w http.ResponseWriter, r *http.Request) {
	requestParams := struct {
		Password string `json:"password"`
		Email    string `json:"email"`
	}{}

	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&requestParams); err != nil {
		respondWithError(w, http.StatusBadRequest, "Could not decode request body", err)
		return
	}

	if requestParams.Password == "" || requestParams.Email == "" {
		respondWithError(w, http.StatusBadRequest, "email and password are required", nil)
		return
	}

	hashedPassword, err := auth.HashPassword(requestParams.Password)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Unable hash the password", err)
		return
	}

	userParams := database.CreateUserParams{
		Email:    requestParams.Email,
		Password: hashedPassword,
	}
	user, err := cfg.db.CreateUser(r.Context(), userParams)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Unable to create user", err)
		return
	}

	responseJSON := User{
		ID:        user.ID,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
		Email:     user.Email,
	}

	respondWithJSON(w, http.StatusCreated, responseJSON)
}
