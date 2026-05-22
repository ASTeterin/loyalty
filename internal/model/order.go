package model

import (
	"context"
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
	ID        string      `json:"id" db:"id"`
	UserID    uuid.UUID   `json:"user_id" db:"user_id"`
	Status    OrderStatus `json:"status" db:"status"`
	CreatedAt time.Time   `json:"created_at" db:"created_at"`
}

type OrderRepository interface {
	Store(ctx context.Context, order Order) error
	GetByOrderID(ctx context.Context, orderID string) (*Order, error)
	ListOrders(ctx context.Context, userID string) ([]Order, error)
}
