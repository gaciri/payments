package gateways

import "payments/models"

type PaymentGateway interface {
	Deposit(models.Transaction) error
	Withdraw(models.Transaction) error
	HandleCallback([]byte) (GateWayResponse, error)
}

type GateWayRequest struct {
	TransactionId string  `json:"transaction_id"`
	Amount        float64 `json:"amount"`
	Currency      string  `json:"currency"`
	CallbackUrl   string  `json:"callback_url"`
	Account       string  `json:"account"`
}
type GateWayResponse struct {
	TransactionId string `json:"transaction_id"`
	Status        string `json:"status"`
}
