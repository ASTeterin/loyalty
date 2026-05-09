package main

import (
	"github.com/gin-gonic/gin"
	"log"
)

func main() {
	r := gin.Default()

	r.POST("/api/user/register", func(c *gin.Context) {
		h.GetShortURL(c, config.ResultBaseURL)
	})
	r.GET("/:id", func(c *gin.Context) {
		h.GetURL(c)
	})
	r.POST("/api/shorten", func(c *gin.Context) {
		restAPIHandler.GetShortURL(c, config.ResultBaseURL)
	})
	r.GET("/ping", func(c *gin.Context) {
		h.CheckDBConnection(c)
	})
	r.POST("/api/shorten/batch", func(c *gin.Context) {
		restAPIHandler.ListShortURLs(c, config.ResultBaseURL)
	})

	if err := r.Run(config.AppAddr); err != nil {
		log.Fatalf("failed to run server: %v", err)
	}
}
