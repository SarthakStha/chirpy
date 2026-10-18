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
	mux.HandleFunc("/healthz", handlerReadiness)
	mux.HandleFunc("/metrics", metricState.handlerMetrics)
	mux.HandleFunc("/reset", metricState.handlerReset)

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
