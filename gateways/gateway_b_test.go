package gateways

import (
	"encoding/xml"
	"github.com/h2non/gock"
	"github.com/stretchr/testify/assert"
	"payments/models"
	"testing"
)

func TestGateWayB_Deposit(t *testing.T) {
	defer gock.Off() // Disable HTTP intercepting after the test

	g := NewGateWayB("http://mock-gateway.com", "http://callback.com")
	transaction := models.Transaction{
		TransactionId: "12345",
		Type:          string(models.Deposit),
		Amount:        100.50,
		Currency:      "USD",
	}

	// Mock the gateway response
	gock.New("http://mock-gateway.com").
		Post("/").
		Reply(202)

	err := g.Deposit(transaction)

	assert.NoError(t, err)
	assert.True(t, gock.IsDone()) // Ensure the mock was called
}
func TestGateWayB_Withdraw(t *testing.T) {
	defer gock.Off()

	g := NewGateWayB("http://mock-gateway.com", "http://callback.com")
	transaction := models.Transaction{
		TransactionId: "54321",
		Type:          string(models.Withdraw),
		Amount:        50.75,
		Currency:      "USD",
	}

	gock.New("http://mock-gateway.com").
		Post("/").
		Reply(202)

	err := g.Withdraw(transaction)

	assert.NoError(t, err)
	assert.True(t, gock.IsDone())
}
func TestGateWayB_HandleCallback(t *testing.T) {
	g := NewGateWayB("http://mock-gateway.com", "http://callback.com")

	respPayload := EnvelopeResponse{
		Body: BodyResponse{
			TransactionResponse: TransactionResponse{
				TransactionId: "12345",
				Status:        "SUCCESS",
			},
		},
	}

	xmlData, err := xml.Marshal(respPayload)
	assert.NoError(t, err)

	callbackResp, err := g.HandleCallback(xmlData)
	assert.NoError(t, err)

	assert.Equal(t, "12345", callbackResp.TransactionId)
	assert.Equal(t, "SUCCESS", callbackResp.Status)
}
func TestGateWayB_Transact_Error(t *testing.T) {
	g := NewGateWayB("http://mock-gateway.com", "http://callback.com")
	transaction := models.Transaction{
		TransactionId: "error-id",
		Type:          string(models.Deposit),
		Amount:        100.50,
		Currency:      "USD",
	}

	// Mock the gateway response with an error
	gock.New("http://mock-gateway.com").
		Post("/").
		Reply(500)

	err := g.Deposit(transaction)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "gateway failed with status code 500")
}
