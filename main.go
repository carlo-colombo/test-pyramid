package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

type HealthResponse struct {
	Alive bool `json:"alive"`
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	resp := HealthResponse{Alive: true}
	json.NewEncoder(w).Encode(resp)
}

func main() {
	http.HandleFunc("/health", healthHandler)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	addr := ":" + port
	server := &http.Server{Addr: addr}

	// Channel to listen for interrupt or terminate signals
	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGTERM)

	go func() {
		log.Printf("Starting server on %s", addr)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("ListenAndServe(): %s", err)
		}
	}()

	<-done // Block until signal is received
	log.Println("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Server Shutdown Failed:%+v", err)
	}
	log.Println("Server exited properly")
}
