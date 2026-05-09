package handler

import (
	"database/sql"
	"errors"
	"github.com/ASTeterin/loyalty/internal/service"
	"github.com/gin-gonic/gin"
	"net/http"
)

type handler struct {
	service service.ShortenerService
	dbConn  *sql.DB
}

type Handler interface {
	Register(c *gin.Context)
}

func NewHandler(service service.ShortenerService, dbConn *sql.DB) Handler {
	return &handler{
		service: service,
		dbConn:  dbConn,
	}
}

func (h *handler) Register(c *gin.Context) {
	login := c.Param("id")
	password := c.PostForm("password")
	userID, err := h.service.Register(login, password)
	if err != nil {
		if errors.Is(err, service.ErrUserExists) {
			c.AbortWithStatus(http.StatusConflict)
			return
		}
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	c.Header("Content-Type", "application/json")
	c.Redirect(http.StatusOK, *userID)
}
