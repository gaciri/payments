package models

import "github.com/go-pg/pg/v10"

type TransactionType string

const (
	Deposit  TransactionType = "deposit"
	Withdraw TransactionType = "withdraw"
)

type TransactionStatus string

const (
	Pending    TransactionStatus = "pending"
	Processing TransactionStatus = "processing"
	Failed     TransactionStatus = "failed"
	Successful TransactionStatus = "successful"
)

type Transaction struct {
	tableName     struct{} `pg:"pay.transactions"`
	TransactionId string   `json:"transaction_id"`
	AccountId     string   `json:"account_id"`
	Type          string   `json:"type"`
	GateWay       string   `json:"gate_way"`
	UserId        string   `json:"user_id"`
	Amount        float64  `json:"amount"`
	Currency      string   `json:"currency"`
	CreatedAt     string   `json:"created_at"`
	UpdatedAt     string   `json:"updated_at"`
	Status        string   `json:"status"`
	RetryCount    int      `json:"retry_count"`
}
type User struct {
	tableName struct{} `pg:"pay.users"`
	Guid      string   `json:"guid"`
	AccountId string   `json:"account_id"`
	GateWay   string   `json:"gate_way"`
	CreatedAt string   `json:"created_at"`
	UpdatedAt string   `json:"updated_at"`
}

func DbGetUser(db *pg.DB, userId string) (User, error) {
	var user User
	err := db.Model(&user).Where("guid = ?", userId).Select()
	return user, err
}
