package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"github.com/SarthakStha/chirpy/internal/auth"
	"github.com/SarthakStha/chirpy/internal/database"
	"github.com/google/uuid"
	"net/http"
	"sort"
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

// Returns all the chirps in the database.
// Takes an optional query parameter author_id, which if provided
// will filter the results for that user.
// Takes an optional query parameter order, sorts by the created date,
// order in ascending by default, sorts in descending order if order=DESC.
// Return Status Codes:
// 200 if no issues.
// 401 if unable parse/invalid user_id.
// 500 if unexpected database issues.
func (cfg *apiConfig) handlerGetAllChirps(w http.ResponseWriter, r *http.Request) {
	queryUser := "author_id"
	userIDString := r.URL.Query().Get(queryUser)

	var queriedChirps []database.Chirp
	var err error

	// If the query parameter is provided will filter on the user provided.
	if userIDString != "" {
		userID, err := uuid.Parse(userIDString)
		if err != nil {
			respondWithError(w, http.StatusUnauthorized, "Unauthorized", err)
		}
		queriedChirps, err = cfg.db.GetUserChirps(r.Context(), userID)
	} else {
		// else all chirps
		queriedChirps, err = cfg.db.GetAllChirps(r.Context())
	}

	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Unable to get chirps", err)
		return
	}

	querySortOrder := "sort"
	order := r.URL.Query().Get(querySortOrder)

	responseChirps := make([]Chirp, len(queriedChirps))
	for i, chirp := range queriedChirps {
		responseChirps[i] = Chirp{
			ID:        chirp.ID,
			CreatedAt: chirp.CreatedAt,
			UpdatedAt: chirp.UpdatedAt,
			Body:      chirp.Body,
			UserID:    chirp.UserID,
		}
	}

	// This works but it is better to do this explicitly
	// Queried Chirps are already sorted by ASC by default
	// Reverse the order if order is DESC
	//if order == "desc" {
	//	slices.Reverse(responseChirps)
	//}
	sort.SliceStable(responseChirps, func(i, j int) bool {
		if order == "desc" {
			return responseChirps[i].CreatedAt.After(responseChirps[j].CreatedAt)
		} else {
			return responseChirps[i].CreatedAt.Before(responseChirps[j].CreatedAt)
		}
	})

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
