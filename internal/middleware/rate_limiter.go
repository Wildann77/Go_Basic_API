package middleware

import (
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	redis "github.com/redis/go-redis/v9"
	limiter "github.com/ulule/limiter/v3"
	mredis "github.com/ulule/limiter/v3/drivers/store/redis"
)

// RateLimiter returns a Gin middleware that limits requests based on IP.
func RateLimiter(client *redis.Client, requests int, period time.Duration) gin.HandlerFunc {
	// 1. Define rate
	rate := limiter.Rate{
		Period: period,
		Limit:  int64(requests),
	}

	// 2. Create Redis store
	store, err := mredis.NewStore(client)
	if err != nil {
		log.Printf("Failed to create rate limiter store: %v", err)
		return func(c *gin.Context) { c.Next() }
	}

	// 3. Create limiter instance
	instance := limiter.New(store, rate)

	return func(c *gin.Context) {
		key := c.ClientIP() // Simple IP-based limiter

		context, err := instance.Get(c, key)
		if err != nil {
			// Fail open on Redis error (log and proceed)
			log.Printf("Rate limiter error: %v", err)
			c.Next()
			return
		}

		c.Header("X-RateLimit-Limit", strconv.FormatInt(context.Limit, 10))
		c.Header("X-RateLimit-Remaining", strconv.FormatInt(context.Remaining, 10))
		c.Header("X-RateLimit-Reset", strconv.FormatInt(context.Reset, 10))

		if context.Reached {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{"error": "Too many requests"})
			return
		}

		c.Next()
	}
}
