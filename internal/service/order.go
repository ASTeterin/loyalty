package service

import (
	"errors"
	"fmt"
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
	CreateOrder(orderID string, userID string) error
	ListOrders(userID string) ([]model.Order, error)
	UserBalance(userID string) (UserBalance, error)
	Withdraw(orderID string, userID string, points float64) error
}

type orderService struct {
	orderRepository         model.OrderRepository
	balanceTransactionsRepo model.BalanceTransactionRepository
}

type UserBalance struct {
	Balance   float64
	Withdrawn float64
}

func NewOrderService(repo model.OrderRepository, balanceTransactionsRepo model.BalanceTransactionRepository) OrderService {
	return &orderService{
		orderRepository:         repo,
		balanceTransactionsRepo: balanceTransactionsRepo,
	}
}

func (s *orderService) CreateOrder(orderID string, userID string) error {
	userUuid, err := uuid.FromString(userID)
	if err != nil {
		return err
	}

	existingOrder, err := s.orderRepository.GetByOrderID(orderID)
	if err != nil {
		if errors.Is(err, model.ErrOrderNotFound) {
			order := model.Order{
				ID:        orderID,
				UserID:    userUuid,
				Status:    model.OrderStatusNew,
				CreatedAt: time.Now(),
			}
			return s.orderRepository.Store(order)
		}
		return err
	}
	if existingOrder.UserID == userUuid {
		return ErrOrderExists
	}
	return ErrOrderCreatedAnotherUser
}

func (s *orderService) ListOrders(userID string) ([]model.Order, error) {
	orders, err := s.orderRepository.ListOrders(userID)
	if err != nil {
		return nil, err
	}
	if len(orders) == 0 {
		return nil, ErrOrdersNotFound
	}
	return orders, nil
}

func (s *orderService) UserBalance(userID string) (UserBalance, error) {
	transactions, err := s.balanceTransactionsRepo.ListUserTransactions(userID)
	fmt.Println("@@@@@@@@@@@@", transactions)
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

func (s *orderService) Withdraw(orderID string, userID string, points float64) error {
	fmt.Println("&##", points)
	err := s.checkBalance(userID, points)
	if err != nil {
		return err
	}

	userUuid, err := uuid.FromString(userID)
	if err != nil {
		return err
	}

	_, err = s.balanceTransactionsRepo.GetByOrderID(orderID)
	fmt.Println("err", err)
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

func (s *orderService) checkBalance(userID string, withdrawPoints float64) error {
	transactions, err := s.balanceTransactionsRepo.ListUserTransactions(userID)
	if err != nil {
		return err
	}
	balance := 0.0
	for _, transaction := range transactions {
		balance += transaction.Points
	}
	fmt.Println(withdrawPoints)
	fmt.Println(balance)
	if balance < withdrawPoints {
		return ErrNotEnoughPoints
	}
	return nil
}
