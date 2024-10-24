package api

import (
	"encoding/json"
	"errors"
	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"github.com/go-chi/chi/v5"
	"github.com/go-pg/pg/v10"
	"github.com/google/uuid"
	log2 "github.com/rs/zerolog/log"
	"io"
	"log"
	"net/http"
	"payments/config"
	"payments/models"
	"payments/utils"
	"time"
)

type Handler struct {
	cfg               *config.Config
	producer          *kafka.Producer
	dbConn            *pg.DB
	availableGateways map[string]bool
}

func NewHandler(cfg *config.Config, producer *kafka.Producer, db *pg.DB, availableGateways map[string]bool) *Handler {
	return &Handler{
		cfg:               cfg,
		producer:          producer,
		dbConn:            db,
		availableGateways: availableGateways,
	}
}

func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	var registerReq RegisterReq
	err := json.NewDecoder(r.Body).Decode(&registerReq)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	_, ok := h.availableGateways[registerReq.GateWay]
	if !ok {
		http.Error(w, "gateway not supported", http.StatusBadRequest)
		return
	}
	userGuid := uuid.NewString()
	user := models.User{
		Guid:      userGuid,
		GateWay:   registerReq.GateWay,
		AccountId: registerReq.AccountId,
		CreatedAt: utils.FmtTimestamp(time.Now()),
	}
	_, err = h.dbConn.Model(&user).Insert()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	resp := RegisterResp{
		UserGuid:  user.Guid,
		GateWay:   user.GateWay,
		AccountId: user.AccountId,
	}
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (h *Handler) Deposit(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	var payRequest PaymentRequest
	err := json.NewDecoder(r.Body).Decode(&payRequest)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	user, err := models.DbGetUser(h.dbConn, payRequest.UserGuid)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	transaction := models.Transaction{
		TransactionId: uuid.New().String(),
		Type:          string(models.Deposit),
		UserId:        payRequest.UserGuid,
		AccountId:     user.AccountId,
		GateWay:       user.GateWay,
		Amount:        payRequest.Amount,
		Currency:      payRequest.Currency,
		CreatedAt:     utils.FmtTimestamp(time.Now()),
		Status:        string(models.Pending),
		RetryCount:    0,
	}
	_, err = h.dbConn.Model(&transaction).Insert()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	jsonData, err := json.Marshal(transaction)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	err = h.producer.Produce(
		&kafka.Message{
			Value: jsonData,
			TopicPartition: kafka.TopicPartition{
				Topic: &h.cfg.KafkaTopics.TransactionTopic, Partition: kafka.PartitionAny,
			},
		}, nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	h.producer.Flush(config.FlushTimeout)
	log2.Info().Str("event", "deposit").Str("transaction_id", transaction.TransactionId).Float64("amount", payRequest.Amount).Str("account", utils.MaskString(transaction.AccountId)).Msg("Transaction received")
	resp := PaymentResponse{
		TransactionId: transaction.TransactionId,
		Amount:        transaction.Amount,
		Currency:      transaction.Currency,
		Status:        transaction.Status,
		Type:          transaction.Type,
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	json.NewEncoder(w).Encode(resp)

}

func (h *Handler) Withdraw(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	var payRequest PaymentRequest
	err := json.NewDecoder(r.Body).Decode(&payRequest)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	user, err := models.DbGetUser(h.dbConn, payRequest.UserGuid)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	transaction := models.Transaction{
		TransactionId: uuid.New().String(),
		Type:          string(models.Withdraw),
		UserId:        payRequest.UserGuid,
		AccountId:     user.AccountId,
		GateWay:       user.GateWay,
		Amount:        payRequest.Amount,
		Currency:      payRequest.Currency,
		CreatedAt:     utils.FmtTimestamp(time.Now()),
		Status:        string(models.Pending),
		RetryCount:    0,
	}
	_, err = h.dbConn.Model(&transaction).Insert()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	jsonData, err := json.Marshal(transaction)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	err = h.producer.Produce(
		&kafka.Message{
			Value: jsonData,
			TopicPartition: kafka.TopicPartition{
				Topic: &h.cfg.KafkaTopics.TransactionTopic, Partition: kafka.PartitionAny,
			},
		}, nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	h.producer.Flush(config.FlushTimeout)
	resp := PaymentResponse{
		TransactionId: transaction.TransactionId,
		Amount:        transaction.Amount,
		Currency:      transaction.Currency,
		Status:        transaction.Status,
		Type:          transaction.Type,
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	json.NewEncoder(w).Encode(resp)
}

func (h *Handler) CheckStatus(w http.ResponseWriter, r *http.Request) {
	transactionId := chi.URLParam(r, "transaction_id")
	var transaction models.Transaction
	err := h.dbConn.Model(&transaction).Where("transaction_id = ?", transactionId).Select()
	if err != nil {
		if errors.Is(err, pg.ErrNoRows) {
			http.Error(w, "transaction not found", http.StatusNotFound)
			return
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
	resp := PaymentResponse{
		TransactionId: transaction.TransactionId,
		Amount:        transaction.Amount,
		Currency:      transaction.Currency,
		Status:        transaction.Status,
		Type:          transaction.Type,
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)

}
func (h *Handler) PaymentCallback(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	transactionId := chi.URLParam(r, "transaction_id")
	log.Printf("Callback invoked for transaction %s", transactionId)
	bytesBody, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	payload := CallbackPayload{
		TransactionId: transactionId,
		Payload:       bytesBody,
	}
	jsonData, err := json.Marshal(payload)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	err = h.producer.Produce(
		&kafka.Message{
			Value: jsonData,
			TopicPartition: kafka.TopicPartition{
				Topic: &h.cfg.KafkaTopics.CallbackTopic, Partition: kafka.PartitionAny,
			},
		}, nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	h.producer.Flush(config.FlushTimeout)

	w.WriteHeader(http.StatusOK)
}
