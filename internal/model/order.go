package model

import (
	"errors"
	"github.com/gofrs/uuid"
	"time"
)

var (
	ErrOrderNotFound = errors.New("order not found")
)

type Order struct {
	ID        int       `json:"id" db:"id"`
	UserID    uuid.UUID `json:"user_id" db:"user_id"`
	Status    string    `json:"status" db:"status"`
	CreatedAt time.Time `json:"pass" db:"created_at"`
}

type OrderRepository interface {
	Store(order Order) error
	GetByOrderID(orderID int) (*Order, error)
}
