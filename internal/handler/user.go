package handler

import (
	"errors"
	"fmt"
	"github.com/ASTeterin/loyalty/internal/cookie"
	"github.com/ASTeterin/loyalty/internal/service"
	"github.com/gin-gonic/gin"
	"net/http"
)

type handler struct {
	service service.UserService
}

type User struct {
	Login    string `json:"login" binding:"required"`
	PassHash string `json:"password" binding:"required"`
}

type Handler interface {
	Register(c *gin.Context)
	Authenticate(c *gin.Context)
}

func NewHandler(service service.UserService) Handler {
	return &handler{
		service: service,
	}
}

func (h *handler) Register(c *gin.Context) {
	body := User{}
	err := c.ShouldBindBodyWithJSON(&body)
	userID, err := h.service.Register(body.Login, body.PassHash)

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
	c.Header("Content-Type", "application/json")
	c.Data(http.StatusOK, "text/plain", []byte(*userID))
}

func (h *handler) Authenticate(c *gin.Context) {
	body := User{}
	err := c.ShouldBindBodyWithJSON(&body)
	userID, err := h.service.Authenticate(body.Login, body.PassHash)

	if err != nil {
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	c.Set(cookie.GetUserKey(), *userID)
	c.Status(http.StatusOK)
}
