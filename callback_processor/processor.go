package callback_processor

import (
	"encoding/json"
	"errors"
	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"github.com/go-pg/pg/v10"
	"github.com/redis/go-redis/v9"
	log2 "github.com/rs/zerolog/log"
	"payments/api"
	"payments/config"
	"payments/gateways"
	"payments/models"
	"payments/utils"
	"time"
)

type CallbackProcessor struct {
	gateWays map[string]gateways.PaymentGateway
	db       *pg.DB
	redisDb  *redis.Client
	cfg      *config.Config
	producer *kafka.Producer
}

func NewCallbackProcessor(cfg *config.Config, db *pg.DB, rdb *redis.Client, producer *kafka.Producer, gateways map[string]gateways.PaymentGateway) *CallbackProcessor {
	return &CallbackProcessor{
		gateWays: gateways,
		db:       db,
		cfg:      cfg,
		redisDb:  rdb,
		producer: producer,
	}
}
func (p CallbackProcessor) Process(payload api.CallbackPayload) error {
	var transaction models.Transaction
	err := p.db.Model(&transaction).Where("transaction_id = ?", payload.TransactionId).Select()
	if err != nil {
		return err
	}
	gateway, ok := p.gateWays[transaction.GateWay]
	if !ok {
		return errors.New("gateway not found")
	}
	resp, err := gateway.HandleCallback(payload.Payload)
	if err != nil {
		return err
	}
	mutexLock := utils.GetMutexLock(p.redisDb, config.TransactionDomain, transaction.TransactionId)
	err = mutexLock.Lock()
	if err != nil {
		return err
	}
	transaction.Status = resp.Status
	transaction.UpdatedAt = utils.FmtTimestamp(time.Now())
	_, err = p.db.Model(&transaction).Where("transaction_id = ?", transaction.TransactionId).Update()
	mutexLock.Unlock()
	if err != nil {
		return err
	}
	jsonData, _ := json.Marshal(transaction)
	err = p.producer.Produce(
		&kafka.Message{
			Value: jsonData,
			TopicPartition: kafka.TopicPartition{
				Topic: &p.cfg.KafkaTopics.TransactionTopic, Partition: kafka.PartitionAny,
			},
		}, nil)
	p.producer.Flush(config.FlushTimeout)
	log2.Info().Str("event", "deposit").Str("transaction_id", transaction.TransactionId).Float64("amount", transaction.Amount).Str("account", utils.MaskString(transaction.AccountId)).Msg("Transaction processed")
	return nil
}
