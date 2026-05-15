package service

import (
	"encoding/json"
	"fmt"
	"github.com/ASTeterin/loyalty/internal/model"
	"io"
	"net/http"
	"strconv"
	"time"
)

const (
	registered string = "REGISTERED"
	invalid    string = "INVALID"
	processing string = "PROCESSING"
	processed  string = "PROCESSED"
)

type OrderResponse struct {
	Order   string `json:"order"`
	Status  string `json:"status"`
	Accrual int    `json:"accrual"`
}

type AccrualService interface {
	ProcessOrder(orderID int) error
}

type accrualService struct {
	baseURL    string
	repo       model.OrderRepository
	httpClient *http.Client
}

func NewAccrualService(repo model.OrderRepository, baseURL string) AccrualService {
	return &accrualService{
		baseURL: baseURL,
		repo:    repo,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (a *accrualService) ProcessOrder(orderID int) error {
	results := make(chan *OrderResponse, 10)
	errors := make(chan error, 10)

	go a.startOrderPolling(orderID, 5*time.Second, 30*time.Second, results, errors)

	for {
		select {
		case order, ok := <-results:
			if !ok {
				return nil
			}
			fmt.Printf("Получен заказ: %s, Статус: %s, Начисление: %d\n",
				order.Order, order.Status, order.Accrual)
			err := a.applyOrderStatus(*order)
			if err != nil {
				errors <- err
			}
		case err, ok := <-errors:
			if !ok {
				continue
			}
			fmt.Printf("Ошибка: %v\n", err)
		}
	}
}

func (a *accrualService) getAccrualPoints(orderNumber int) (*OrderResponse, error) {
	url := fmt.Sprintf("%s/api/orders/%d", a.baseURL, orderNumber)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("не удалось создать запрос: %w", err)
	}

	fmt.Println(req)
	resp, err := a.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("ошибка HTTP-запроса: %w", err)
	}
	defer resp.Body.Close()

	fmt.Println(resp)
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

func (a *accrualService) startOrderPolling(orderID int, interval, duration time.Duration, results chan<- *OrderResponse, errors chan<- error) {
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

func (a *accrualService) applyOrderStatus(order OrderResponse) error {
	fmt.Println(order)
	orderID, err := strconv.Atoi(order.Order)
	if err != nil {
		return err
	}
	orderData, err := a.repo.GetByOrderID(orderID)
	if err != nil {
		return err
	}

	status, err := convertOrderStatus(order.Status)
	fmt.Println("stat", status)
	if err != nil {
		return err
	}
	orderData.Status = status

	return a.repo.Store(*orderData)
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
