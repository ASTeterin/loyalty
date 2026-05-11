package handler

import (
	"errors"
	"fmt"
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
	fmt.Println(err)
	fmt.Println(userID)
	if err != nil {
		if errors.Is(err, service.ErrUserExists) {
			c.AbortWithStatus(http.StatusConflict)
			return
		}
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	c.Header("Content-Type", "application/json")
	c.Data(http.StatusOK, "text/plain", []byte(*userID))
}
