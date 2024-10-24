package main

import (
	"context"
	"errors"
	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"log"
	"net/http"
	"os"
	"os/signal"
	"payments/api"
	"payments/config"
	"payments/utils"
	"syscall"
	"time"
)

func main() {
	router := chi.NewRouter()
	router.Use(middleware.RequestID)
	router.Use(middleware.Logger)
	router.Use(middleware.Recoverer)
	cfg, err := config.ReadConfig()
	if err != nil {
		log.Fatalf("failed to read config file %v", err)
	}
	producer, err := kafka.NewProducer(&kafka.ConfigMap{"bootstrap.servers": cfg.Kafka.Server})
	if err != nil {
		log.Fatalf("failed to create kafka producer: %v", err)
	}
	defer producer.Close()
	dbConn := utils.NewDbConnection(cfg)
	availableGateways := map[string]bool{"a": true, "b": true}
	handler := api.NewHandler(cfg, producer, dbConn, availableGateways)
	router.Post("/register", handler.Register)
	router.Post("/deposit", handler.Deposit)
	router.Post("/withdraw", handler.Withdraw)
	router.Get("/status/{transaction_id}", handler.CheckStatus)
	router.Post("/callback/{transaction_id}", handler.PaymentCallback)

	srv := &http.Server{
		Addr:    ":8080",
		Handler: router,
	}
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	go func() {
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("Server error: %v", err)
		}
	}()
	log.Println("Server is running on :8080")
	<-stop
	log.Println("Shutting down gracefully...")
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server stopped")
}
