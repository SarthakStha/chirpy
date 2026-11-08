package main

import (
	"encoding/json"
	"github.com/SarthakStha/chirpy/internal/auth"
	"net/http"
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

	responseJSON := User{
		ID:        user.ID,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
		Email:     user.Email,
	}
	respondWithJSON(w, http.StatusOK, responseJSON)
}
