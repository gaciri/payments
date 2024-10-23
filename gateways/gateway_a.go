package gateways

import (
	"bytes"
	"encoding/json"
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

type GateWayA struct {
	gateWayDomain  string
	withdrawPath   string
	depositPath    string
	callbackPrefix string
}

func NewGateWayA(gateWayDomain, withdrawPath, depositPath, callbackPrefix string) *GateWayA {
	return &GateWayA{
		gateWayDomain:  gateWayDomain,
		withdrawPath:   withdrawPath,
		depositPath:    depositPath,
		callbackPrefix: callbackPrefix,
	}
}

func (g *GateWayA) Deposit(transaction models.Transaction) error {
	err := g.transact(transaction)
	return err
}

func (g *GateWayA) Withdraw(transaction models.Transaction) error {
	err := g.transact(transaction)
	return err
}

func (g *GateWayA) transact(transaction models.Transaction) error {
	defer gock.Off() //
	var path string
	switch transaction.Type {
	case string(models.Deposit):
		path = g.depositPath
	case string(models.Withdraw):
		path = g.withdrawPath
	}
	url := g.gateWayDomain + path
	url, err := utils.JoinUrlPaths(g.gateWayDomain, path)
	if err != nil {
		return err
	}
	callbackUrl, err := utils.JoinUrlPaths(g.callbackPrefix, transaction.TransactionId)
	if err != nil {
		return err
	}
	// we simulate a third of failures
	rChoice := rand.Intn(3)
	if rChoice == 2 { // failure path
		rStatus := rand.Intn(5) + 500
		gock.New(g.gateWayDomain).Post(path).Reply(rStatus)
	} else {
		defer func() {

			rSeconds := rand.Intn(5)
			var status string
			rChoice := rand.Intn(3)
			if rChoice == 2 {
				status = string(models.Failed)
			} else {
				status = string(models.Successful)
			}
			resp := GateWayResponse{
				TransactionId: transaction.TransactionId,
				Status:        status,
			}
			time.AfterFunc(time.Second*time.Duration(rSeconds), func() {
				jsonData, err := json.Marshal(resp)
				if err != nil {
					log.Printf("Error marshalling json %s", err)
					return
				}
				req, err := http.NewRequest(http.MethodPost, callbackUrl, bytes.NewBuffer(jsonData))
				if err != nil {
					log.Printf("error generating request %v", err)
				}
				_, err = http.DefaultClient.Do(req)
				if err != nil {
					log.Printf("error making request %s", err)
				}

			})

		}()
		gock.New(g.gateWayDomain).Post(path).Reply(202)

	}
	gateWayReq := GateWayRequest{
		TransactionId: transaction.TransactionId,
		Amount:        transaction.Amount,
		Currency:      transaction.Currency,
		CallbackUrl:   callbackUrl,
		Account:       transaction.AccountId,
	}
	jsonData, err := json.Marshal(gateWayReq)
	if err != nil {
		return err
	}
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(jsonData))
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

func (g *GateWayA) HandleCallback(payload []byte) (GateWayResponse, error) {
	var resp GateWayResponse
	err := json.Unmarshal(payload, &resp)
	return resp, err
}
