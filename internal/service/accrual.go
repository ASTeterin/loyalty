package service

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/ASTeterin/loyalty/internal/model"
	"github.com/gofrs/uuid"
	"io"
	"net/http"
	"time"
)

const (
	registered string = "REGISTERED"
	invalid    string = "INVALID"
	processing string = "PROCESSING"
	processed  string = "PROCESSED"
)

type OrderResponse struct {
	Order   string  `json:"order"`
	Status  string  `json:"status"`
	Accrual float64 `json:"accrual"`
}

type AccrualService interface {
	ProcessOrder(ctx context.Context, orderID string, userID string) error
}

type accrualService struct {
	baseURL     string
	orderRepo   model.OrderRepository
	balanceRepo model.BalanceTransactionRepository
	httpClient  *http.Client
}

func NewAccrualService(repo model.OrderRepository, baseURL string, balanceRepo model.BalanceTransactionRepository) AccrualService {
	return &accrualService{
		baseURL:     baseURL,
		orderRepo:   repo,
		balanceRepo: balanceRepo,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (a *accrualService) ProcessOrder(ctx context.Context, orderID string, userID string) error {
	results := make(chan *OrderResponse, 10)
	errors := make(chan error, 10)

	go a.startOrderPolling(orderID, 5*time.Second, 30*time.Second, results, errors)

	for {
		select {
		case order, ok := <-results:
			if !ok {
				return nil
			}

			err := a.applyOrderStatus(ctx, orderID, order.Status)
			if err != nil {
				errors <- err
			}
			err = a.applyBalanceTransaction(orderID, order.Accrual, userID)
			if err != nil {
				errors <- err
			}
		case err, ok := <-errors:
			if !ok {
				continue
			}
			fmt.Printf("Error: %v\n", err)
		}
	}
}

func (a *accrualService) getAccrualPoints(orderNumber string) (*OrderResponse, error) {
	url := fmt.Sprintf("%s/api/orders/%s", a.baseURL, orderNumber)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("не удалось создать запрос: %w", err)
	}

	resp, err := a.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("ошибка HTTP-запроса: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("неудачный статус ответа: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("ошибка чтения ответа: %w", err)
	}

	var order OrderResponse
	if err := json.Unmarshal(body, &order); err != nil {
		return nil, fmt.Errorf("ошибка парсинга JSON: %w", err)
	}

	return &order, nil
}

func (a *accrualService) startOrderPolling(orderID string, interval, duration time.Duration, results chan<- *OrderResponse, errors chan<- error) {
	defer close(results)
	defer close(errors)

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	done := time.After(duration)

	for {
		select {
		case <-done:
			return
		case <-ticker.C:
			order, err := a.getAccrualPoints(orderID)
			if err != nil {
				errors <- err
				continue
			}
			results <- order
		}
	}
}

func (a *accrualService) applyOrderStatus(ctx context.Context, orderID string, orderStatus string) error {
	orderData, err := a.orderRepo.GetByOrderID(ctx, orderID)
	if err != nil {
		return err
	}

	status, err := convertOrderStatus(orderStatus)
	if err != nil {
		return err
	}
	orderData.Status = status
	return a.orderRepo.Store(ctx, *orderData)
}

func (a *accrualService) applyBalanceTransaction(orderID string, accrual float64, userID string) error {
	userUid, err := uuid.FromString(userID)
	if err != nil {
		return err
	}

	balanceTransaction := model.BalanceTransaction{
		OrderID: orderID,
		UserID:  userUid,
		Points:  accrual,
	}
	return a.balanceRepo.Store(balanceTransaction)
}

func convertOrderStatus(orderStatus string) (model.OrderStatus, error) {
	switch orderStatus {
	case registered, processing:
		return model.OrderStatusProcessing, nil
	case processed:
		return model.OrderStatusProcessed, nil
	case invalid:
		return model.OrderStatusInvalid, nil
	default:
		return model.OrderStatusInvalid, fmt.Errorf("invalid order status %s", orderStatus)
	}
}
