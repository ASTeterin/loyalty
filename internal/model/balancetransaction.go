package model

import (
	"github.com/gofrs/uuid"
	"time"
)

type BalanceTransaction struct {
	OrderID   int       `json:"status" db:"order_id"`
	UserID    uuid.UUID `json:"user_id" db:"user_id"`
	Points    float64   `json:"points" db:"points"`
	CreatedAt time.Time `json:"created_ar" db:"created_at"`
}

type BalanceTransactionRepository interface {
	ListUserTransactions(userID string) ([]BalanceTransaction, error)
	Store(t BalanceTransaction) error
}
