package main

import (
	"encoding/xml"
	"log"
	"payments/gateways"
)

func main() {
	withdrawReq := gateways.WithdrawEnvelope{
		SoapEnv: "http://schemas.xmlsoap.org/soap/envelope/",
		Web:     "http://example.com/",
		Body: gateways.WithdrawBody{
			Withdraw: gateways.WithdrawRequest{
				Account:       "234556780987",
				TransactionID: "xxxxxxxxx",
				Amount:        100,
				Currency:      "USD",
			},
		},
	}
	req, err := xml.MarshalIndent(withdrawReq, "", "  ")
	if err != nil {
		panic(err)
	}
	soapReq := append([]byte(xml.Header), req...)
	log.Println(string(soapReq))
}
