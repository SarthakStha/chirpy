package main

import (
	"fmt"
	"github.com/SarthakStha/chirpy/internal/database"
	"log"
	"net/http"
	"sync/atomic"
)

type apiConfig struct {
	fileserverHits atomic.Int32
	db             *database.Queries
	platform       string
}

func (cfg *apiConfig) middlewareMetricInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cfg.fileserverHits.Add(1)
		next.ServeHTTP(w, r)
	})
}

func (cfg *apiConfig) handlerMetrics(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)

	htmlTemplate := `
<html>
  <body>
    <h1>Welcome, Chirpy Admin</h1>
    <p>Chirpy has been visited %d times!</p>
  </body>
</html>
	`
	body := fmt.Sprintf(htmlTemplate, cfg.fileserverHits.Load())
	if _, err := w.Write([]byte(body)); err != nil {
		log.Printf("error writing response: %s", err)
	}
}

func (cfg *apiConfig) handlerReset(w http.ResponseWriter, r *http.Request) {
	if cfg.platform != "dev" {
		respondWithError(w, http.StatusUnauthorized, "Reset can only be ran in dev environment", fmt.Errorf("attempted to ran reset in non-dev environgment"))
	}

	cfg.fileserverHits.And(0)
	if err := cfg.db.ResetUsers(r.Context()); err != nil {
		respondWithError(w, http.StatusInternalServerError, "Unable to reset users table", err)
		return
	}

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	if _, err := w.Write([]byte("Hits reset to 0\nDeleted all users\n")); err != nil {
		log.Printf("error writing response: %s", err)
	}
}
