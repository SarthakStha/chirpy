package main

import (
	"database/sql"
	"encoding/json"
	"errors"
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

type NewUserParams struct {
	Password string `json:"password"`
	Email    string `json:"email"`
}

func (cfg *apiConfig) handlerNewUser(w http.ResponseWriter, r *http.Request) {
	requestParams := NewUserParams{}

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

// Updates the user's email and password with the ones provided in the body.
// Extracts the user information from the JWT in Authentication header.
// Response Status Codes:
// 200 if the user credentials are updated without any issues.
// 400 if unable to parse the request body.
// 401 if the Auth Header is invalid/misformatted, user_id doesn't exist in the database.
// 500 if unable to hash the password or unexpexted database connection error.
func (cfg *apiConfig) handlerUpdateUser(w http.ResponseWriter, r *http.Request) {
	authToken, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Unauthorized", err)
		return
	}

	user_id, err := auth.ValidateJWT(authToken, cfg.secret)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Unauthorized", err)
		return
	}

	requestParams := NewUserParams{}
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&requestParams); err != nil {
		respondWithError(w, http.StatusBadRequest, "Could not decode request body", err)
		return
	}

	hashedPassword, err := auth.HashPassword(requestParams.Password)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Internal Server Error", err)
		return
	}

	updatedUserParams := database.UpdateUserParams{
		Email:    requestParams.Email,
		Password: hashedPassword,
		ID:       user_id,
	}

	updatedUser, err := cfg.db.UpdateUser(r.Context(), updatedUserParams)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			respondWithError(w, http.StatusUnauthorized, "Unauthorized", err)
			return
		}

		respondWithError(w, http.StatusInternalServerError, "Internal Server Error", err)
		return
	}

	responseJSON := User{
		ID:        updatedUser.ID,
		CreatedAt: updatedUser.CreatedAt,
		UpdatedAt: updatedUser.UpdatedAt,
		Email:     updatedUser.Email,
	}

	respondWithJSON(w, http.StatusOK, responseJSON)
}
