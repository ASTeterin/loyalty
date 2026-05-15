package model

import (
	"errors"
	"github.com/gofrs/uuid"
	"time"
)

var (
	ErrOrderNotFound = errors.New("order not found")
)

type OrderStatus string

const (
	OrderStatusNew        OrderStatus = "NEW"
	OrderStatusProcessing OrderStatus = "PROCESSING"
	OrderStatusInvalid    OrderStatus = "INVALID"
	OrderStatusProcessed  OrderStatus = "PROCESSED"
)

type Order struct {
	ID        int         `json:"id" db:"id"`
	UserID    uuid.UUID   `json:"user_id" db:"user_id"`
	Status    OrderStatus `json:"status" db:"status"`
	CreatedAt time.Time   `json:"pass" db:"created_at"`
}

type OrderRepository interface {
	Store(order Order) error
	GetByOrderID(orderID int) (*Order, error)
	ListOrders(userID string) ([]Order, error)
}
