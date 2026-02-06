package handlers

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"time"
)

func HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":    "healthy",
		"timestamp": time.Now().Unix(),
		"service":   "goapi",
		"version":   "1.0.0",
	})
}