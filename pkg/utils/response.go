package utils

import (
	"fmt"

	"github.com/gin-gonic/gin"
)

type Response struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
	Error   interface{} `json:"error,omitempty"`
	Meta    *Meta       `json:"meta,omitempty"`
}

type Meta struct {
	Page  int `json:"page,omitempty"`
	Limit int `json:"limit,omitempty"`
	Total int `json:"total,omitempty"`
}

func SuccessResponse(c *gin.Context, status int, message string, data interface{}) {
	c.JSON(status, Response{
		Success: true,
		Message: message,
		Data:    data,
	})
}

func ErrorResponse(c *gin.Context, status int, message string, err interface{}) {
	if err != nil {
		// Attach error to context for logging middleware
		var e error
		switch v := err.(type) {
		case error:
			e = v
		case string:
			e = fmt.Errorf("%s", v)
		default:
			e = fmt.Errorf("%v", v)
		}
		_ = c.Error(e) // Add to Gin errors
	}

	c.JSON(status, Response{
		Success: false,
		Message: message,
		Error:   err,
	})
}

func PaginatedResponse(c *gin.Context, status int, message string, data interface{}, page, limit, total int) {
	c.JSON(status, Response{
		Success: true,
		Message: message,
		Data:    data,
		Meta: &Meta{
			Page:  page,
			Limit: limit,
			Total: total,
		},
	})
}
