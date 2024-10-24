package api

type RegisterReq struct {
	GateWay   string `json:"gate_way"`
	AccountId string `json:"account_id"`
}
type RegisterResp struct {
	UserGuid  string `json:"user_guid"`
	GateWay   string `json:"gate_way"`
	AccountId string `json:"account_id"`
}

type PaymentRequest struct {
	Amount         float64 `json:"amount"`
	Currency       string  `json:"currency"`
	UserGuid       string  `json:"user_guid"`
	ClientCallback string  `json:"callback"`
}
type PaymentResponse struct {
	TransactionId string  `json:"transaction_id"`
	Amount        float64 `json:"amount"`
	Currency      string  `json:"currency"`
	Status        string  `json:"status"`
	Type          string  `json:"type"`
}
type CallbackPayload struct {
	TransactionId string `json:"transaction_id"`
	Payload       []byte `json:"payload"`
}
