package callback_processor

import (
	"errors"
	"github.com/go-pg/pg/v10"
	"github.com/redis/go-redis/v9"
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
}

func NewCallbackProcessor(cfg *config.Config, db *pg.DB, rdb *redis.Client, gateways map[string]gateways.PaymentGateway) *CallbackProcessor {
	return &CallbackProcessor{
		gateWays: gateways,
		db:       db,
		cfg:      cfg,
		redisDb:  rdb,
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
	return nil
}
