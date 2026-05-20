package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gorilla/mux"

	"backend-assignment/internal/handlers"
	"backend-assignment/internal/middleware"
	"backend-assignment/internal/services"
	"backend-assignment/internal/store"
)

func main() {
	
	rateStore := store.NewRateStore(5, time.Minute)
	productStore := store.NewProductStore()

	
	rateHandler := handlers.NewRateHandler(
		services.NewRateService(rateStore),
	)

	productHandler := handlers.NewProductHandler(
		services.NewProductService(productStore),
	)

	// create Route
	router := mux.NewRouter()

	// Middleware
	router.Use(middleware.Logging)

	// Root / Health Route
	router.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		w.WriteHeader(http.StatusOK)

		w.Write([]byte(`{
		"message":"Backend Assignment API Running",
		"status":"healthy"
		}`))
	}).Methods(http.MethodGet)

	// Optional dedicated health endpoint
	router.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		w.WriteHeader(http.StatusOK)

		w.Write([]byte(`{
			"status":"ok"
		}`))
	}).Methods(http.MethodGet)

	// Rate limiter routes
	router.HandleFunc("/request", rateHandler.CreateRequest).
		Methods(http.MethodPost)

	router.HandleFunc("/stats", rateHandler.Stats).
		Methods(http.MethodGet)

	// Product routes
	router.HandleFunc("/products", productHandler.Create).
		Methods(http.MethodPost)

	router.HandleFunc("/products", productHandler.List).
		Methods(http.MethodGet)

	router.HandleFunc("/products/{id}", productHandler.GetByID).
		Methods(http.MethodGet)

	router.HandleFunc("/products/{id}/media", productHandler.AppendMedia).
		Methods(http.MethodPost)

	// Dynamic PORT for Render deployment
	port := os.Getenv("PORT")

	if port == "" {
		port = "8080"
	}

	// HTTP server configuration
	server := &http.Server{
		Addr:              ":" + port,
		Handler:           router,
		ReadHeaderTimeout: 5 * time.Second,
	}

	// Start server
	go func() {
		log.Printf("server listening on port %s", port)

		if err := server.ListenAndServe(); err != nil &&
			err != http.ErrServerClosed {

			log.Fatalf("server error: %v", err)
		}
	}()

	// Graceful shutdown handling
	quit := make(chan os.Signal, 1)

	signal.Notify(
		quit,
		os.Interrupt,
		syscall.SIGTERM,
	)

	<-quit

	log.Println("shutdown signal received")

	ctx, cancel := context.WithTimeout(
		context.Background(),
		5*time.Second,
	)

	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("server shutdown error: %v", err)
	}

	log.Println("server stopped gracefully")
}
