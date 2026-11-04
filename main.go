package main

import (
	"database/sql"
	"fmt"
	"github.com/SarthakStha/chirpy/internal/database"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"log"
	"net/http"
	"os"
	"time"
)

func main() {
	const port = "8080"
	const rootPath = "."

	// Loading .env file
	if err := godotenv.Load(); err != nil {
		log.Fatal("error loading .env file: ", err)
	}

	// Checking if the env variables have values
	dbURL := os.Getenv("DB_URL")
	if dbURL == "" {
		log.Fatal("db connection string is empty")
	}

	platform := os.Getenv("PLATFORM")
	if platform == "" {
		log.Fatal("platform must be set")
	}

	// Connecting to the database
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatal("error connecting to the database: ", err)
	}
	dbQueries := database.New(db)

	mux := http.NewServeMux()
	fileServer := http.FileServer(http.Dir(rootPath))
	cfg := &apiConfig{
		db:       dbQueries,
		platform: platform,
	}

	mux.Handle("/app/", cfg.middlewareMetricInc(http.StripPrefix("/app", fileServer)))

	mux.HandleFunc("GET /api/healthz", handlerReadiness)

	mux.HandleFunc("GET /admin/metrics", cfg.handlerMetrics)
	mux.HandleFunc("POST /admin/reset", cfg.handlerReset)

	mux.HandleFunc("POST /api/users", cfg.handlerNewUser)
	mux.HandleFunc("POST /api/chirps", cfg.handlerNewChirp)
	mux.HandleFunc("GET /api/chirps", cfg.handlerGetAllChirps)
	mux.HandleFunc("GET /api/chirps/{chirpID}", cfg.handlerGetChirp)

	server := &http.Server{
		Addr:           ":" + port,
		Handler:        mux,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	fmt.Println("Server is now running. Listening to port:", port)
	log.Fatal(server.ListenAndServe())
}
