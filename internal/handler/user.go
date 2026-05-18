package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/ASTeterin/loyalty/internal/cookie"
	"github.com/ASTeterin/loyalty/internal/service"
	"github.com/gin-gonic/gin"
	"net/http"
	"strconv"
	"time"
	"unicode"
)

type handler struct {
	userService    service.UserService
	orderService   service.OrderService
	accrualService service.AccrualService
}

type User struct {
	Login    string `json:"login" binding:"required"`
	PassHash string `json:"password" binding:"required"`
}

type Order struct {
	Number     int    `json:"number"`
	Status     string `json:"status"`
	Accrual    int    `json:"accrual"`
	UploadedAt string `json:"uploaded_at"`
}

type UserBalance struct {
	Current   float64 `json:"current"`
	Withdrawn float64 `json:"withdrawn"`
}

type Handler interface {
	Register(c *gin.Context)
	Authenticate(c *gin.Context)
	CreateOrder(c *gin.Context)
	ListOrders(c *gin.Context)
	GetUserBalance(c *gin.Context)
}

func NewHandler(userService service.UserService, orderService service.OrderService, accrualService service.AccrualService) Handler {
	return &handler{
		userService:    userService,
		orderService:   orderService,
		accrualService: accrualService,
	}
}

func (h *handler) Register(c *gin.Context) {
	body := User{}
	err := c.ShouldBindBodyWithJSON(&body)
	userID, err := h.userService.Register(body.Login, body.PassHash)

	if err != nil {
		fmt.Println(err)
		if errors.Is(err, service.ErrUserExists) {
			c.AbortWithStatus(http.StatusConflict)
			return
		}
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	c.Set(cookie.GetUserKey(), *userID)
	c.Status(http.StatusOK)
}

func (h *handler) Authenticate(c *gin.Context) {
	body := User{}
	err := c.ShouldBindBodyWithJSON(&body)
	userID, err := h.userService.Authenticate(body.Login, body.PassHash)

	if err != nil {
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	c.Set(cookie.GetUserKey(), *userID)
	c.Status(http.StatusOK)
}

func (h *handler) CreateOrder(c *gin.Context) {
	fmt.Println("CRE!!!!!!!!!!!!")
	var orderID string
	err := c.BindPlain(&orderID)
	if err != nil {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}
	if validateOrder(orderID) != nil {
		c.AbortWithStatus(http.StatusUnprocessableEntity)
		return
	}
	intOrderID, err := strconv.Atoi(orderID)
	if err != nil {
		c.AbortWithStatus(http.StatusBadRequest)
	}

	userID := getUserID(c)
	go func() {
		err2 := h.accrualService.ProcessOrder(intOrderID, userID)
		if err2 != nil {
			// TODO: add logging
			fmt.Println("accrual points err:", err2)
		}
	}()

	err = h.orderService.CreateOrder(intOrderID, userID)
	if err != nil {
		if errors.Is(err, service.ErrOrderExists) {
			c.Status(http.StatusOK)
			return
		}
		if errors.Is(err, service.ErrOrderCreatedAnotherUser) {
			c.AbortWithStatus(http.StatusConflict)
		}
		c.AbortWithStatus(http.StatusInternalServerError)
	}

	c.Status(http.StatusAccepted)
}

func (h *handler) ListOrders(c *gin.Context) {
	userID := getUserID(c)
	orders, err := h.orderService.ListOrders(userID)
	if err != nil {
		if errors.Is(err, service.ErrOrdersNotFound) {
			c.Status(http.StatusNoContent)
			return
		}
		c.AbortWithStatus(http.StatusInternalServerError)
	}

	responseData := make([]Order, 0, len(orders))
	for _, order := range orders {
		responseData = append(responseData, Order{
			Number:     order.ID,
			Status:     string(order.Status),
			UploadedAt: order.CreatedAt.Format(time.RFC3339),
		})
	}

	response, err := json.Marshal(responseData)
	if err != nil {
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}
	c.Data(http.StatusOK, "application/json", response)
}

func (h *handler) GetUserBalance(c *gin.Context) {
	userID := getUserID(c)
	balance, err := h.orderService.UserBalance(userID)
	if err != nil {
		c.AbortWithStatus(http.StatusInternalServerError)
	}

	response, err := json.Marshal(UserBalance{
		Current:   balance.Balance,
		Withdrawn: balance.Withdrawn,
	})
	if err != nil {
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}
	c.Data(http.StatusOK, "application/json", response)
}

func validateOrder(orderID string) error {
	if !isDigitsOnly(orderID) {
		return fmt.Errorf("order number contains more than just numbers: %s", orderID)
	}
	if !validateLuhn(orderID) {
		return fmt.Errorf("order number is not valid: %s", orderID)
	}
	return nil
}

func isDigitsOnly(s string) bool {
	for _, r := range s {
		if !unicode.IsDigit(r) {
			return false
		}
	}
	return len(s) > 0
}

func validateLuhn(number string) bool {
	sum := 0
	alternating := false
	for i := len(number) - 1; i >= 0; i-- {
		digit := int(number[i] - '0')

		if alternating {
			digit *= 2
			if digit > 9 {
				digit -= 9
			}
		}

		sum += digit
		alternating = !alternating
	}

	return sum%10 == 0
}

func getUserID(c *gin.Context) string {
	return c.GetString(cookie.GetUserKey())
}
