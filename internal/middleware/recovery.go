package middleware

import (
	"fmt"
	"goapi/pkg/logger"
	"net"
	"net/http"
	"net/http/httputil"
	"os"
	"runtime/debug"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// CustomRecovery is a middleware that recovers from any panics and writes a 500 if there was one.
func CustomRecovery() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				// Check for a broken connection, as it is not really a
				// condition that warrants a panic stack trace.
				var brokenPipe bool
				if ne, ok := err.(*net.OpError); ok {
					if se, ok := ne.Err.(*os.SyscallError); ok {
						if strings.Contains(strings.ToLower(se.Error()), "broken pipe") || strings.Contains(strings.ToLower(se.Error()), "connection reset by peer") {
							brokenPipe = true
						}
					}
				}

				stack := debug.Stack()
				httpRequest, _ := httputil.DumpRequest(c.Request, false)
				if brokenPipe {
					logger.Error("Panic Recovered (Broken Pipe)",
						"error", err,
						"request", httpRequest,
					)
					// If the connection is dead, we can't write a status to it.
					c.Error(err.(error)) // nolint: errcheck
					c.Abort()
					return
				}

				logger.Error("Panic Recovered",
					"error", err,
					"stack", string(stack),
					"path", c.Request.URL.Path,
					"request_id", c.GetString(RequestIDKey),
				)

				c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
					"error":      "Internal Server Error",
					"message":    fmt.Sprintf("Panic: %v", err),
					"request_id": c.GetString(RequestIDKey),
					"timestamp":  time.Now().Format(time.RFC3339),
				})
			}
		}()
		c.Next()
	}
}
