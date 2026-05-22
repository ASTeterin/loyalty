package service

import (
	"context"
	"errors"
	"github.com/ASTeterin/loyalty/internal/model"
	"github.com/gofrs/uuid"
	"math"
	"time"
)

var (
	ErrOrderExists             = errors.New("order already exists")
	ErrOrderCreatedAnotherUser = errors.New("order created by another user")
	ErrOrdersNotFound          = errors.New("orders not found user")
	ErrNotEnoughPoints         = errors.New("not enough points for withdraw")
)

type OrderService interface {
	CreateOrder(ctx context.Context, orderID string, userID string) error
	ListOrders(ctx context.Context, userID string) ([]model.Order, error)
	UserBalance(ctx context.Context, userID string) (UserBalance, error)
	Withdraw(ctx context.Context, orderID string, userID string, points float64) error
	ListWithdrawals(ctx context.Context, userID string) ([]WithdrawalData, error)
}

type orderService struct {
	orderRepository         model.OrderRepository
	balanceTransactionsRepo model.BalanceTransactionRepository
}

type UserBalance struct {
	Balance   float64
	Withdrawn float64
}

type WithdrawalData struct {
	OrderID     string
	Points      float64
	ProcessedAt time.Time
}

func NewOrderService(repo model.OrderRepository, balanceTransactionsRepo model.BalanceTransactionRepository) OrderService {
	return &orderService{
		orderRepository:         repo,
		balanceTransactionsRepo: balanceTransactionsRepo,
	}
}

func (s *orderService) CreateOrder(ctx context.Context, orderID string, userID string) error {
	userUuid, err := uuid.FromString(userID)
	if err != nil {
		return err
	}

	existingOrder, err := s.orderRepository.GetByOrderID(ctx, orderID)
	if err != nil {
		if errors.Is(err, model.ErrOrderNotFound) {
			order := model.Order{
				ID:        orderID,
				UserID:    userUuid,
				Status:    model.OrderStatusNew,
				CreatedAt: time.Now(),
			}
			return s.orderRepository.Store(ctx, order)
		}
		return err
	}
	if existingOrder.UserID == userUuid {
		return ErrOrderExists
	}
	return ErrOrderCreatedAnotherUser
}

func (s *orderService) ListOrders(ctx context.Context, userID string) ([]model.Order, error) {
	orders, err := s.orderRepository.ListOrders(ctx, userID)
	if err != nil {
		return nil, err
	}
	if len(orders) == 0 {
		return nil, ErrOrdersNotFound
	}
	return orders, nil
}

func (s *orderService) UserBalance(ctx context.Context, userID string) (UserBalance, error) {
	transactions, err := s.balanceTransactionsRepo.ListUserTransactions(ctx, userID, false)
	if err != nil {
		return UserBalance{}, err
	}
	balance := UserBalance{
		Balance:   0,
		Withdrawn: 0,
	}
	for _, transaction := range transactions {
		balance.Balance += transaction.Points
		if transaction.Points < 0 {
			balance.Withdrawn += math.Abs(transaction.Points)
		}
	}

	return balance, nil
}

func (s *orderService) Withdraw(ctx context.Context, orderID string, userID string, points float64) error {
	err := s.checkBalance(ctx, userID, points)
	if err != nil {
		return err
	}

	userUuid, err := uuid.FromString(userID)
	if err != nil {
		return err
	}

	_, err = s.balanceTransactionsRepo.GetByOrderID(orderID)
	if err != nil {
		if errors.Is(err, model.ErrBalanceTransactionNotFound) {
			t := model.BalanceTransaction{
				OrderID:   orderID,
				UserID:    userUuid,
				Points:    -points,
				CreatedAt: time.Now(),
			}
			return s.balanceTransactionsRepo.Store(t)
		}
		return err
	}

	return nil
}

func (s *orderService) ListWithdrawals(ctx context.Context, userID string) ([]WithdrawalData, error) {
	transactions, err := s.balanceTransactionsRepo.ListUserTransactions(ctx, userID, true)
	if err != nil {
		return nil, err
	}
	result := make([]WithdrawalData, 0, len(transactions))
	for _, transaction := range transactions {
		result = append(result, WithdrawalData{
			OrderID:     transaction.OrderID,
			Points:      -transaction.Points,
			ProcessedAt: transaction.CreatedAt,
		})
	}
	return result, nil
}

func (s *orderService) checkBalance(ctx context.Context, userID string, withdrawPoints float64) error {
	transactions, err := s.balanceTransactionsRepo.ListUserTransactions(ctx, userID, false)
	if err != nil {
		return err
	}
	balance := 0.0
	for _, transaction := range transactions {
		balance += transaction.Points
	}
	if balance < withdrawPoints {
		return ErrNotEnoughPoints
	}
	return nil
}
