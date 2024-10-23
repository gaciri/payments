package payment_processor

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/afex/hystrix-go/hystrix"
	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"github.com/go-pg/pg/v10"
	"github.com/redis/go-redis/v9"
	"log"
	"payments/config"
	"payments/gateways"
	"payments/models"
	"payments/utils"
	"time"
)

type PaymentProcessor struct {
	gateWays map[string]gateways.PaymentGateway
	db       *pg.DB
	redisDb  *redis.Client
	producer *kafka.Producer
	cfg      *config.Config
}

func NewPaymentProcessor(cfg *config.Config, db *pg.DB, rdb *redis.Client, producer *kafka.Producer, gateways map[string]gateways.PaymentGateway) *PaymentProcessor {
	for key, _ := range gateways {
		hystrix.ConfigureCommand(
			key,
			hystrix.CommandConfig{
				Timeout:               config.GateWayTimeout,
				MaxConcurrentRequests: config.MaxConcurrentRequests,
				ErrorPercentThreshold: config.ErrorPercentThreshold,
				SleepWindow:           config.CircuitBreakSleepWindow,
			},
		)
	}
	return &PaymentProcessor{
		gateWays: gateways,
		db:       db,
		cfg:      cfg,
		producer: producer,
		redisDb:  rdb,
	}
}

func (p *PaymentProcessor) Process(payload models.Transaction) error {
	var transaction models.Transaction
	err := p.db.Model(&transaction).Where("transaction_id = ?", payload.TransactionId).Select()
	if err != nil {
		return err
	}
	if transaction.Status != string(models.Pending) {
		return nil
	}
	gateway, ok := p.gateWays[transaction.GateWay]
	if !ok {
		return errors.New(fmt.Sprintf("Payment Gateway not found for gate id: %s", transaction.GateWay))
	}
	mutexLock := utils.GetMutexLock(p.redisDb, config.TransactionDomain, transaction.TransactionId)
	err = mutexLock.Lock()
	if err != nil {
		return err
	}
	transaction.Status = string(models.Processing)
	transaction.UpdatedAt = utils.FmtTimestamp(time.Now())
	transaction.RetryCount = transaction.RetryCount + 1
	_, err = p.db.Model(&transaction).Where("transaction_id = ?", transaction.TransactionId).Update()
	mutexLock.Unlock()
	if err != nil {
		return err
	}
	hystrix.Go(transaction.GateWay, func() error {
		switch transaction.Type {
		case string(models.Deposit):
			err := gateway.Deposit(transaction)
			if err != nil {
				log.Printf("Calling getway deposit error: %s", err)
				return err
			}

		case string(models.Withdraw):
			err := gateway.Withdraw(transaction)
			if err != nil {
				log.Printf("Calling getway withdraw error: %s", err)
				return err
			}
		}
		return nil
	}, func(err error) error {
		if transaction.RetryCount < config.MaxGateWayRetries {
			duration := utils.ExponentialBackoff(transaction.RetryCount)
			mutexLock := utils.GetMutexLock(p.redisDb, config.TransactionDomain, transaction.TransactionId)
			err = mutexLock.Lock()
			if err != nil {
				return err
			}
			transaction.Status = string(models.Pending)
			_, dbErr := p.db.Model(&transaction).Where("transaction_id = ?", transaction.TransactionId).Update()
			mutexLock.Unlock()
			if dbErr != nil {
				return dbErr
			}

			time.AfterFunc(duration, func() {
				jsonData, _ := json.Marshal(transaction)
				err = p.producer.Produce(
					&kafka.Message{
						Value: jsonData,
						TopicPartition: kafka.TopicPartition{
							Topic: &p.cfg.KafkaTopics.TransactionTopic, Partition: kafka.PartitionAny,
						},
					}, nil)
			})
		} else {
			mutexLock := utils.GetMutexLock(p.redisDb, config.TransactionDomain, transaction.TransactionId)
			err = mutexLock.Lock()
			transaction.Status = string(models.Failed)
			_, dbErr := p.db.Model(&transaction).Where("transaction_id = ?", transaction.TransactionId).Update()
			mutexLock.Unlock()
			if dbErr != nil {
				return dbErr
			}
		}
		return nil
	})
	return nil
}
