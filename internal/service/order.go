package service

import (
	"errors"
	"github.com/ASTeterin/loyalty/internal/model"
	"github.com/gofrs/uuid"
	"time"
)

var (
	ErrOrderExists             = errors.New("order already exists")
	ErrOrderCreatedAnotherUser = errors.New("order created by another user")
	ErrOrdersNotFound          = errors.New("orders not found user")
)

type OrderService interface {
	CreateOrder(orderID int, userID string) error
	ListOrders(userID string) ([]model.Order, error)
}

type orderService struct {
	repo model.OrderRepository
}

func NewOrderService(repo model.OrderRepository) OrderService {
	return &orderService{
		repo: repo,
	}
}

func (s *orderService) CreateOrder(orderID int, userID string) error {
	userUuid, err := uuid.FromString(userID)
	if err != nil {
		return err
	}

	existingOrder, err := s.repo.GetByOrderID(orderID)
	if err != nil {
		if errors.Is(err, model.ErrOrderNotFound) {
			order := model.Order{
				ID:        orderID,
				UserID:    userUuid,
				Status:    model.OrderStatusNew,
				CreatedAt: time.Now(),
			}
			return s.repo.Store(order)
		}
		return err
	}
	if existingOrder.UserID == userUuid {
		return ErrOrderExists
	}
	return ErrOrderCreatedAnotherUser
}

func (s *orderService) ListOrders(userID string) ([]model.Order, error) {
	orders, err := s.repo.ListOrders(userID)
	if err != nil {
		return nil, err
	}
	if len(orders) == 0 {
		return nil, ErrOrdersNotFound
	}
	return orders, nil
}
