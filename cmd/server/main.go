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

	rateHandler := handlers.NewRateHandler(services.NewRateService(rateStore))
	productHandler := handlers.NewProductHandler(services.NewProductService(productStore))

	router := mux.NewRouter()
	router.Use(middleware.Logging)

	router.HandleFunc("/request", rateHandler.CreateRequest).Methods(http.MethodPost)
	router.HandleFunc("/stats", rateHandler.Stats).Methods(http.MethodGet)
	router.HandleFunc("/products", productHandler.Create).Methods(http.MethodPost)
	router.HandleFunc("/products", productHandler.List).Methods(http.MethodGet)
	router.HandleFunc("/products/{id}", productHandler.GetByID).Methods(http.MethodGet)
	router.HandleFunc("/products/{id}/media", productHandler.AppendMedia).Methods(http.MethodPost)

	server := &http.Server{
		Addr:              ":8080",
		Handler:           router,
		ReadHeaderTimeout: 5 * time.Second,
	}

	go func() {
		log.Println("server listening on http://localhost:8080")
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server error: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("server shutdown error: %v", err)
	}
	log.Println("server stopped")
}
