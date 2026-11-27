package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"github.com/SarthakStha/chirpy/internal/auth"
	"github.com/google/uuid"
	"net/http"
)

// Checks for the event "user.upgraded".
// Sets is_chirpy_red to True for the user_id provided.
// Authenticates the request using the API key provided in the header.
// Return Status Codes:
// 204 if everything is successfull or in case any other event is provided.
// 400 if unable to decode request body/user_id from the body or invalid/missing
// Authentication header.
// 401 if the provided API key is invalid.
// 404 if the user is not found in the database.
// 500 if there is an unexpected error with the database.
func (cfg *apiConfig) handlerPremiumUser(w http.ResponseWriter, r *http.Request) {
	APIKey, err := auth.GetAPIKey(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Unauthorized", err)
		return
	}

	if cfg.polkaKey != APIKey {
		respondWithError(w, http.StatusUnauthorized, "Unauthorized", err)
		return
	}

	requestParams := struct {
		Event string `json:"event"`
		Data  struct {
			UserID string `json:"user_id"`
		} `json:"data"`
	}{}

	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&requestParams); err != nil {
		respondWithError(w, http.StatusBadRequest, "Bad Request", err)
		return
	}

	if requestParams.Event != "user.upgraded" {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	// Decoder can be used to directly parse into UUID, worth looking into the behaviour of
	// Unmarshall on non primitive types.
	userID, err := uuid.Parse(requestParams.Data.UserID)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Bad Request", err)
		return
	}

	if err := cfg.db.AddPremium(r.Context(), userID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		respondWithError(w, http.StatusInternalServerError, "Internal Server Error", err)
	}

	w.WriteHeader(http.StatusNoContent)
}
