package main

import (
	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"net/http"
	"payments/api"
	"payments/config"
	"payments/utils"
)

func main() {
	router := chi.NewRouter()
	router.Use(middleware.RequestID)
	router.Use(middleware.Logger)
	router.Use(middleware.Recoverer)
	cfg, err := config.ReadConfig()
	if err != nil {
		panic(err)
	}
	producer, err := kafka.NewProducer(&kafka.ConfigMap{"bootstrap.servers": cfg.Kafka.Server})
	if err != nil {
		panic(err)
	}
	defer producer.Close()
	dbConn := utils.NewDbConnection(cfg)
	handler := api.NewHandler(cfg, producer, dbConn)
	router.Post("/register", handler.Register)
	router.Post("/deposit", handler.Deposit)
	router.Post("/withdraw", handler.Withdraw)
	router.Get("/status/{transaction_id}", handler.CheckStatus)
	router.Post("/callback/{transaction_id}", handler.PaymentCallback)
	http.ListenAndServe(":8080", router)
}
