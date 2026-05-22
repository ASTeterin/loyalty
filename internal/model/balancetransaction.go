package model

import (
	"context"
	"errors"
	"github.com/gofrs/uuid"
	"time"
)

var ErrBalanceTransactionNotFound = errors.New("balance transaction not found")

type BalanceTransaction struct {
	OrderID   string    `json:"order_id" db:"order_id"`
	UserID    uuid.UUID `json:"user_id" db:"user_id"`
	Points    float64   `json:"points" db:"points"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

type BalanceTransactionRepository interface {
	ListUserTransactions(ctx context.Context, userID string, onlyWithdrawal bool) ([]BalanceTransaction, error)
	Store(t BalanceTransaction) error
	GetByOrderID(orderID string) (*BalanceTransaction, error)
}
