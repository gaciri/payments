package gateways

import (
	"github.com/h2non/gock"
	"github.com/stretchr/testify/assert"
	"payments/models"
	"testing"
)

func TestNewGateWayA(t *testing.T) {
	gateWay := NewGateWayA("https://gateway.example.com", "/withdraw", "/deposit", "https://callback.example.com")
	assert.NotNil(t, gateWay)
	assert.Equal(t, "https://gateway.example.com", gateWay.gateWayDomain)
	assert.Equal(t, "/withdraw", gateWay.withdrawPath)
	assert.Equal(t, "/deposit", gateWay.depositPath)
	assert.Equal(t, "https://callback.example.com", gateWay.callbackPrefix)
}

func TestGateWayA_Deposit_Success(t *testing.T) {
	gock.Off() // Clean up previous mocks
	defer gock.Off()

	transaction := models.Transaction{
		TransactionId: "12345",
		Amount:        100.0,
		Currency:      "USD",
		Type:          string(models.Deposit),
	}

	gock.New("https://gateway.example.com").
		Post("/deposit").
		Reply(202) // HTTP 202 Accepted

	gock.New("https://callback.example.com").
		Post("12345").
		Reply(200) // Simulating callback success

	gateWay := NewGateWayA("https://gateway.example.com", "/withdraw", "/deposit", "https://callback.example.com")
	err := gateWay.Deposit(transaction)

	assert.NoError(t, err)
}
func TestGateWayA_Withdraw_Success(t *testing.T) {
	gock.Off() // Clean up previous mocks
	defer gock.Off()

	// Setting up mock HTTP response
	transaction := models.Transaction{
		TransactionId: "12345",
		Amount:        100.0,
		Currency:      "USD",
		Type:          string(models.Withdraw),
	}

	gock.New("https://gateway.example.com").
		Post("/withdraw").
		Reply(202) // HTTP 202 Accepted

	gock.New("https://callback.example.com").
		Post("12345").
		Reply(200) // Simulating callback success

	gateWay := NewGateWayA("https://gateway.example.com", "/withdraw", "/deposit", "https://callback.example.com")
	err := gateWay.Withdraw(transaction)

	assert.NoError(t, err)
}

func TestGateWayA_Deposit_Failure(t *testing.T) {
	gock.Off() // Clean up previous mocks
	defer gock.Off()

	// Setting up mock HTTP response for failure
	transaction := models.Transaction{
		TransactionId: "12345",
		Amount:        100.0,
		Currency:      "USD",
		Type:          string(models.Deposit),
	}

	gock.New("https://gateway.example.com").
		Post("/deposit").
		Reply(500) // Simulate an internal server error

	gateWay := NewGateWayA("https://gateway.example.com", "/withdraw", "/deposit", "https://callback.example.com")
	err := gateWay.Deposit(transaction)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "gateway failed with status code 500")
}
func TestGateWayA_Withdraw_Failure(t *testing.T) {
	gock.Off() // Clean up previous mocks
	defer gock.Off()

	// Setting up mock HTTP response for failure
	transaction := models.Transaction{
		TransactionId: "12345",
		Amount:        100.0,
		Currency:      "USD",
		Type:          string(models.Withdraw),
	}

	gock.New("https://gateway.example.com").
		Post("/withdraw").
		Reply(500) // Simulate an internal server error

	gateWay := NewGateWayA("https://gateway.example.com", "/withdraw", "/deposit", "https://callback.example.com")
	err := gateWay.Withdraw(transaction)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "gateway failed with status code 500")
}
