package main

import (
	"encoding/json"
	"errors"
	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"log"
	"payments/config"
	"payments/gateways"
	"payments/models"
	"payments/payment_processor"
	"payments/utils"
)

func main() {
	cfg, err := config.ReadConfig()
	if err != nil {
		panic(err)
	}
	consumer, err := utils.NewConsumer(cfg.Kafka.Server, "transactions")
	if err != nil {
		panic(err)
	}
	err = consumer.Subscribe(cfg.KafkaTopics.TransactionTopic, nil)
	if err != nil {
		panic(err)
	}
	producer, err := kafka.NewProducer(&kafka.ConfigMap{"bootstrap.servers": cfg.Kafka.Server})
	if err != nil {
		panic(err)
	}
	db := utils.NewDbConnection(cfg)
	rdb := utils.NewRedisConnection(cfg)
	gateWays := map[string]gateways.PaymentGateway{
		"a": gateways.NewGateWayA(cfg.Network.GateWayAUrl, "/withdraw", "/deposit", cfg.Network.CallbackPrefix),
		"b": gateways.NewGateWayB(cfg.Network.GateWayBUrl, cfg.Network.CallbackPrefix),
	}

	processor := payment_processor.NewPaymentProcessor(cfg, db, rdb, producer, gateWays)
	for {
		msg, err := consumer.ReadMessage(-1)

		if err != nil {
			var kafkaError kafka.Error
			if errors.As(err, &kafkaError) && kafkaError.Code() == kafka.ErrTimedOut {
				continue // No message received, keep polling
			}
			log.Printf("Error reading message from Kafka topic %s: %v", cfg.KafkaTopics.TransactionTopic, err)
			continue
		}
		log.Printf("Message on %s: %s\n", msg.TopicPartition, string(msg.Value))
		var transaction models.Transaction
		err = json.Unmarshal(msg.Value, &transaction)
		if err != nil {
			log.Printf("Error unmarshalling transaction: %v", err)
			continue
		}
		err = processor.Process(transaction) // process sequentially, launch goroutine downstream
		if err != nil {
			log.Printf("Error processing transaction: %s: %v", transaction.TransactionId, err)
		}

	}
}
