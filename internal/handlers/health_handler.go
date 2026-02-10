package handlers

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

type HealthHandler struct {
	db    *gorm.DB
	redis *redis.Client
}

func NewHealthHandler(db *gorm.DB, redis *redis.Client) *HealthHandler {
	return &HealthHandler{db: db, redis: redis}
}

func (h *HealthHandler) Check(c *gin.Context) {
	status := "healthy"
	components := make(map[string]string)

	// Check DB
	sqlDB, err := h.db.DB()
	if err != nil {
		status = "unhealthy"
		components["db"] = "failed to get instance"
	} else if err := sqlDB.Ping(); err != nil {
		status = "unhealthy"
		components["db"] = "ping failed: " + err.Error()
	} else {
		components["db"] = "up"
	}

	// Check Redis
	if err := h.redis.Ping(context.Background()).Err(); err != nil {
		status = "unhealthy"
		components["redis"] = "ping failed: " + err.Error()
	} else {
		components["redis"] = "up"
	}

	statusCode := http.StatusOK
	if status == "unhealthy" {
		statusCode = http.StatusServiceUnavailable
	}

	c.JSON(statusCode, gin.H{
		"status":     status,
		"timestamp":  time.Now().Unix(),
		"service":    "goapi",
		"version":    "1.0.0",
		"components": components,
	})
}
