package gateways

import (
	"bytes"
	"encoding/xml"
	"errors"
	"fmt"
	"github.com/h2non/gock"
	"log"
	"math/rand"
	"net/http"
	"payments/models"
	"payments/utils"
	"time"
)

type WithdrawEnvelope struct {
	XMLName xml.Name     `xml:"soapenv:Envelope"`
	SoapEnv string       `xml:"xmlns:soapenv,attr"`
	Web     string       `xml:"xmlns:web,attr"`
	Body    WithdrawBody `xml:"soapenv:Body"`
}
type WithdrawBody struct {
	Withdraw WithdrawRequest `xml:"web:Withdraw"`
}
type WithdrawRequest struct {
	Account       string  `xml:"xmlns:account,attr"`
	TransactionID string  `xml:"xmlns:transaction,attr"`
	Amount        float64 `xml:"xmlns:amount,attr"`
	Currency      string  `xml:"xmlns:currency,attr"`
	Callback      string  `xml:"xmlns:callback,attr"`
}

type DepositEnvelope struct {
	XMLName xml.Name    `xml:"soapenv:Envelope"`
	SoapEnv string      `xml:"xmlns:soapenv,attr"`
	Web     string      `xml:"xmlns:web,attr"`
	Body    DepositBody `xml:"soapenv:Body"`
}

type DepositBody struct {
	Deposit DepositRequest `xml:"web:Deposit"`
}

type DepositRequest struct {
	Account       string  `xml:"xmlns:account,attr"`
	TransactionID string  `xml:"xmlns:transaction,attr"`
	Amount        float64 `xml:"xmlns:amount,attr"`
	Currency      string  `xml:"xmlns:currency,attr"`
	Callback      string  `xml:"xmlns:callback,attr"`
}
type EnvelopeResponse struct {
	XMLName xml.Name `xml:"Envelope"`
	Body    BodyResponse
}

type BodyResponse struct {
	TransactionResponse TransactionResponse `xml:"TransactionResponse"`
}
type TransactionResponse struct {
	TransactionId string `xml:"TransactionId"`
	Status        string `xml:"Status"`
}

type GateWayB struct {
	gateWayUrl     string
	callbackPrefix string
}

func NewGateWayB(gateWayUrl, callbackPrefix string) *GateWayB {
	return &GateWayB{
		gateWayUrl:     gateWayUrl,
		callbackPrefix: callbackPrefix,
	}
}

func (g *GateWayB) transact(transaction models.Transaction) error {
	defer gock.Off()
	var payload []byte
	callbackUrl, err := utils.JoinUrlPaths(g.callbackPrefix, transaction.TransactionId)
	if err != nil {
		return err
	}
	switch transaction.Type {
	case string(models.Deposit):
		depositEnvelope := DepositEnvelope{
			SoapEnv: "http://schemas.xmlsoap.org/soap/envelope/",
			Web:     g.gateWayUrl,
			Body: DepositBody{
				Deposit: DepositRequest{
					Account:  transaction.AccountId,
					Amount:   transaction.Amount,
					Currency: transaction.Currency,
					Callback: callbackUrl,
				},
			},
		}
		payload, err = xml.Marshal(depositEnvelope)
		if err != nil {
			return err
		}

	case string(models.Withdraw):
		withdrawEnvelope := WithdrawEnvelope{
			SoapEnv: "http://schemas.xmlsoap.org/soap/envelope/",
			Web:     g.gateWayUrl,
			Body: WithdrawBody{
				Withdraw: WithdrawRequest{
					Account:  transaction.AccountId,
					Amount:   transaction.Amount,
					Currency: transaction.Currency,
					Callback: callbackUrl,
				},
			},
		}
		payload, err = xml.Marshal(withdrawEnvelope)
		if err != nil {
			return err
		}
	}
	soapReq := append([]byte(xml.Header), payload...)
	// we simulate a third of failures
	rChoice := rand.Intn(3)
	if rChoice == 2 {
		rStatus := rand.Intn(5) + 500
		gock.New(g.gateWayUrl).Post("/").Reply(rStatus)
	} else {
		defer func() {

			rSeconds := rand.Intn(5) // simulate callback delay
			var status string
			rChoice := rand.Intn(3)
			if rChoice == 2 {
				status = string(models.Failed)
			} else {
				status = string(models.Successful)
			}

			respPayload := EnvelopeResponse{
				Body: BodyResponse{
					TransactionResponse: TransactionResponse{
						TransactionId: transaction.TransactionId,
						Status:        status,
					},
				},
			}
			time.AfterFunc(time.Second*time.Duration(rSeconds), func() {
				xmlData, err := xml.Marshal(respPayload)
				if err != nil {
					log.Printf("Error marshalling json %s", err)
					return
				}
				xmlData = append([]byte(xml.Header), xmlData...)
				req, err := http.NewRequest(http.MethodPost, callbackUrl, bytes.NewBuffer(xmlData))
				if err != nil {
					log.Printf("error generating request %v", err)
				}
				_, err = http.DefaultClient.Do(req)
				if err != nil {
					log.Printf("error making request %s", err)
				}

			})

		}()
		gock.New(g.gateWayUrl).Post("/").Reply(202)
	}
	req, err := http.NewRequest(http.MethodPost, g.gateWayUrl, bytes.NewBuffer(soapReq))
	if err != nil {
		return err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	// we might want to isolate client errors (4xx)
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return errors.New(fmt.Sprintf("gateway failed with status code %d", resp.StatusCode))
	}
	return nil
}

func (g *GateWayB) Deposit(transaction models.Transaction) error {
	err := g.transact(transaction)
	return err
}
func (g *GateWayB) Withdraw(transaction models.Transaction) error {
	err := g.transact(transaction)
	return err
}

func (g *GateWayB) HandleCallback(payload []byte) (GateWayResponse, error) {
	var resp EnvelopeResponse
	var gateWayResponse GateWayResponse
	err := xml.Unmarshal(payload, &resp)
	if err != nil {
		return gateWayResponse, err
	}
	gateWayResponse = GateWayResponse{
		TransactionId: resp.Body.TransactionResponse.TransactionId,
		Status:        resp.Body.TransactionResponse.Status,
	}
	return gateWayResponse, nil
}
