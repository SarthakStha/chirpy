package main

import (
	"fmt"
	"log"
	"net/http"
	"time"
)

func main() {
	const port = "8080"
	const rootPath = "."

	mux := http.NewServeMux()
	fileServer := http.FileServer(http.Dir(rootPath))
	metricState := &apiConfig{}

	mux.Handle("/app/", metricState.middlewareMetricInc(http.StripPrefix("/app", fileServer)))

	mux.HandleFunc("GET /api/healthz", handlerReadiness)
	mux.HandleFunc("POST /api/validate_chirp", handlerValidate)

	mux.HandleFunc("GET /admin/metrics", metricState.handlerMetrics)
	mux.HandleFunc("POST /admin/reset", metricState.handlerReset)

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
