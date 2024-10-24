package main

import (
	"encoding/json"
	"errors"
	"github.com/afex/hystrix-go/hystrix"
	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"log"
	"payments/callback_dispatcher"
	"payments/config"
	"payments/models"
	"payments/utils"
)

func main() {
	cfg, err := config.ReadConfig()
	if err != nil {
		panic(err)
	}
	consumer, err := utils.NewConsumer(cfg.Kafka.Server, config.DispatcherConsumerGroup)
	if err != nil {
		panic(err)
	}
	err = consumer.Subscribe(cfg.KafkaTopics.DispatcherTopic, nil)
	if err != nil {
		panic(err)
	}
	processor := callback_dispatcher.NewCallbackDispatcher()
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
		var callbackPayload models.Transaction
		err = json.Unmarshal(msg.Value, &callbackPayload)
		if err != nil {
			log.Printf("Error unmarshalling transaction: %v", err)
			continue
		}
		if callbackPayload.ClientCallback == "" {
			continue
		}
		domain, err := utils.ExtractDomain(callbackPayload.ClientCallback)
		if err != nil {
			log.Printf("Error extracting domain: %v", err)
			continue
		}
		// set up circuit breaker to client domain
		hystrix.ConfigureCommand(
			domain,
			hystrix.CommandConfig{
				Timeout:               config.GateWayTimeout,
				MaxConcurrentRequests: config.MaxConcurrentRequests,
				ErrorPercentThreshold: config.ErrorPercentThreshold,
				SleepWindow:           config.CircuitBreakSleepWindow,
			},
		)
		hystrix.Go(domain, func() error {
			err := processor.Process(callbackPayload)
			return err
		}, nil)

	}
}
