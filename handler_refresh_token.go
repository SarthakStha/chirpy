package main

import (
	"database/sql"
	"errors"
	"github.com/SarthakStha/chirpy/internal/auth"
	"net/http"
	"time"
)

// Reads the refresh token on the Authentication header and return a new JWT.
// Response JSON : {"Token": <new JWT Token>}
// return codes:
// 200 if everything works as expected.
// 500 if there is some unexpected connection issues with the database.
// 401 if the header is missing/misformatted or if the token doesn't exist or is invalid
func (cfg *apiConfig) handlerRefreshJWT(w http.ResponseWriter, r *http.Request) {
	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Unauthorized", err)
		return
	}

	refreshTokenEntry, err := cfg.db.GetRefreshTokenEntry(r.Context(), token)
	if err != nil {
		// If the db responds with no rows for the given refresh token then it's a invalid refresh
		// token not a server issue
		if errors.Is(err, sql.ErrNoRows) {
			respondWithError(w, http.StatusUnauthorized, "Unauthorized", errors.New("Attempted to refresh JWT with a Refresh Token which doesn't exist"))
			return
		}

		respondWithError(w, http.StatusInternalServerError, "Internal Server Error", err)
		return
	}

	if refreshTokenEntry.RevokedAt.Valid || time.Now().UTC().After(refreshTokenEntry.ExpiresAt) {
		respondWithError(w, http.StatusUnauthorized, "Unauthorized", errors.New("The refresh token has already expired."))
		return
	}

	// Default JWT duration is 1 hour.
	JWTDuration := time.Hour * 1
	tokenString, err := auth.MakeJWT(refreshTokenEntry.UserID, cfg.secret, JWTDuration)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Internal Server Error", err)
		return
	}

	responseJSON := struct {
		Token string `json:"token"`
	}{
		Token: tokenString,
	}
	respondWithJSON(w, http.StatusOK, responseJSON)
}

// Revokes the refresh token provided in the Authentication header.
// Will only throw an error if the header is missing or if there are unexpected
// db connection issues. It will silently pass in the case of invalid refresh tokens.
func (cfg *apiConfig) handlerRevokeRefreshToken(w http.ResponseWriter, r *http.Request) {
	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Missing/Invalid Auth Credential", err)
		return
	}

	// Will only update entries that are not already revoked i.e. revoked_at is NULL
	_, err = cfg.db.RevokeRefreshToken(r.Context(), token)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		respondWithError(w, http.StatusInternalServerError, "Unable to revoke the token", err)
		return
	}

	// respondWithJSON is not required here since we are returning status code 204 no content
	//respondWithJSON(w, http.StatusNoContent, struct{}{})
	w.WriteHeader(http.StatusNoContent)
}
