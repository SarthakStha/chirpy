package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"github.com/SarthakStha/chirpy/internal/auth"
	"github.com/SarthakStha/chirpy/internal/database"
	"github.com/google/uuid"
	"net/http"
	"strings"
	"time"
)

type Chirp struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Body      string    `json:"body"`
	UserID    uuid.UUID `json:"user_id"`
}

func (cfg *apiConfig) handlerNewChirp(w http.ResponseWriter, r *http.Request) {
	authHeader, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Missing/Invalid Auth Credential", err)
		return
	}

	userID, err := auth.ValidateJWT(authHeader, cfg.secret)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Unauthorized User", errors.New("Invalid JWT"))
		return
	}

	requestParams := struct {
		Body string `json:"body"`
	}{}
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&requestParams); err != nil {
		respondWithError(w, http.StatusBadRequest, "Could not decode request body", err)
		return
	}

	cleanedChirp, err := validateChirp(requestParams.Body)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error(), err)
	}

	chirpArgs := database.CreateChirpParams{
		Body:   cleanedChirp,
		UserID: userID,
	}
	chirp, err := cfg.db.CreateChirp(r.Context(), chirpArgs)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Unable to create chirp", err)
		return
	}

	responseJSON := Chirp{
		ID:        chirp.ID,
		CreatedAt: chirp.CreatedAt,
		UpdatedAt: chirp.UpdatedAt,
		Body:      chirp.Body,
		UserID:    chirp.UserID,
	}
	respondWithJSON(w, http.StatusCreated, responseJSON)
}

func validateChirp(body string) (string, error) {
	const charLimit = 140
	if len(body) > charLimit {
		return "", errors.New("Chirp is too long")
	}

	profaneWords := map[string]struct{}{
		"kerfuffle": {},
		"sharbert":  {},
		"fornax":    {},
	}
	return cleanChirp(body, profaneWords), nil
}

func cleanChirp(body string, profaneWords map[string]struct{}) string {
	words := strings.Split(body, " ")
	for i, word := range words {
		if _, ok := profaneWords[strings.ToLower(word)]; ok {
			words[i] = "****"
		}
	}
	cleanedBody := strings.Join(words, " ")
	return cleanedBody
}

func (cfg *apiConfig) handlerGetAllChirps(w http.ResponseWriter, r *http.Request) {
	allChirps, err := cfg.db.GetAllChirps(r.Context())
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Unable to get chirps", err)
		return
	}

	responseChirps := make([]Chirp, len(allChirps))
	for i, chirp := range allChirps {
		responseChirps[i] = Chirp{
			ID:        chirp.ID,
			CreatedAt: chirp.CreatedAt,
			UpdatedAt: chirp.UpdatedAt,
			Body:      chirp.Body,
			UserID:    chirp.UserID,
		}
	}

	respondWithJSON(w, http.StatusOK, responseChirps)
}

func (cfg *apiConfig) handlerGetChirp(w http.ResponseWriter, r *http.Request) {
	chirpID := r.PathValue("chirpID")
	chirpUUID, err := uuid.Parse(chirpID)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Unable to parse chirp ID provided", err)
		return
	}

	chirp, err := cfg.db.GetChirp(r.Context(), chirpUUID)
	if err != nil {
		respondWithError(w, http.StatusNotFound, "Unable to find chirp", err)
		return
	}

	responseChirp := Chirp{
		ID:        chirp.ID,
		CreatedAt: chirp.CreatedAt,
		UpdatedAt: chirp.UpdatedAt,
		Body:      chirp.Body,
		UserID:    chirp.UserID,
	}
	respondWithJSON(w, http.StatusOK, responseChirp)
}

// Attempts to delete the chirp provided in the request path.
// Gets the user_id from the Authorization header.
// Verifies if the user owns the chirps, if yes deletes it.
// Return Status Codes:
// 204 if everything is successfull, the chirp is deleted from the database.
// 400 unable to parse the chip_id from the request path.
// 401 invalid/missing Auth header.
// 403 attempting to delete other user's chirp.
// 404 chirp not found in the database.
// 500 unexpected database error.
func (cfg *apiConfig) handlerDeleteChirp(w http.ResponseWriter, r *http.Request) {
	authToken, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Unauthorized", err)
		return
	}

	userID, err := auth.ValidateJWT(authToken, cfg.secret)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Unauthorized", err)
		return
	}

	chirpID, err := uuid.Parse(r.PathValue("chirpID"))
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Bad Request", err)
		return
	}

	chirpEntry, err := cfg.db.GetChirp(r.Context(), chirpID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			respondWithError(w, http.StatusNotFound, "Not Found", err)
			return
		}
		respondWithError(w, http.StatusInternalServerError, "Internal Server Error", err)
		return
	}

	if chirpEntry.UserID != userID {
		respondWithError(w, http.StatusForbidden, "Unauthorized", err)
		return
	}

	err = cfg.db.DeleteChirp(r.Context(), chirpID)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Internal Server Error", err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
