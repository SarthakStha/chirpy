package main

import (
	"encoding/json"
	"github.com/SarthakStha/chirpy/internal/auth"
	"github.com/SarthakStha/chirpy/internal/database"
	"github.com/google/uuid"
	"net/http"
	"time"
)

func (cfg *apiConfig) handlerLogin(w http.ResponseWriter, r *http.Request) {
	type params struct {
		Password string `json:"password"`
		Email    string `json:"email"`
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

	// Default JWT duration is 1 hour.
	JWTDuration := time.Hour * 1
	tokenString, err := auth.MakeJWT(user.ID, cfg.secret, JWTDuration)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Unable to create a JWT for the user", err)
		return
	}

	// Default refresh token duration is 60 days.
	refreshTokenDuration := time.Hour * 24 * 60
	refreshToken := auth.MakeRefreshToken()
	refreshTokenParams := database.CreateRefreshTokenParams{
		Token:     refreshToken,
		UserID:    user.ID,
		ExpiresAt: time.Now().UTC().Add(refreshTokenDuration),
	}
	_, err = cfg.db.CreateRefreshToken(r.Context(), refreshTokenParams)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Unable to store new refresh token", err)
		return
	}

	responseJSON := struct {
		ID           uuid.UUID `json:"id"`
		CreatedAt    time.Time `json:"created_at"`
		UpdatedAt    time.Time `json:"updated_at"`
		Email        string    `json:"email"`
		Token        string    `json:"token"`
		RefreshToken string    `json:"refresh_token"`
	}{
		ID:           user.ID,
		CreatedAt:    user.CreatedAt,
		UpdatedAt:    user.UpdatedAt,
		Email:        user.Email,
		Token:        tokenString,
		RefreshToken: refreshToken,
	}
	respondWithJSON(w, http.StatusOK, responseJSON)
}
