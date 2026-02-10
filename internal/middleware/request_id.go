package middleware

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
)

const RequestIDKey = "RequestID"

func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Check for incoming header
		requestID := c.GetHeader("X-Request-ID")
		if requestID == "" {
			requestID = generateRequestID()
		}

		// Set in context and header
		c.Set(RequestIDKey, requestID)
		c.Writer.Header().Set("X-Request-ID", requestID)

		c.Next()
	}
}

func generateRequestID() string {
	// Simple random hex string + timestamp for reasonable uniqueness
	b := make([]byte, 12)
	_, err := rand.Read(b)
	if err != nil {
		// Fallback if random fails (unlikely)
		return fmt.Sprintf("%d", time.Now().UnixNano())
	}
	return hex.EncodeToString(b)
}
