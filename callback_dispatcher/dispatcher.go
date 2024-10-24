package callback_dispatcher

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"payments/api"
	"payments/models"
)

type CallbackDispatcher struct {
}

func NewCallbackDispatcher() *CallbackDispatcher {
	return &CallbackDispatcher{}
}

func (d *CallbackDispatcher) Process(transaction models.Transaction) error {
	paymentResp := api.PaymentResponse{
		TransactionId: transaction.TransactionId,
		Amount:        transaction.Amount,
		Currency:      transaction.Currency,
		Status:        transaction.Status,
		Type:          transaction.Type,
	}
	jsonData, err := json.Marshal(paymentResp)
	if err != nil {
		return err
	}
	req, err := http.NewRequest(http.MethodPost, transaction.ClientCallback, bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	// we might want to isolate client errors (4xx)
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return errors.New(fmt.Sprintf("client failed with status code %d", resp.StatusCode))
	}
	return nil
}
