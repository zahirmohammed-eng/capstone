package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

func SecurityAndAuth(apiKey string) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("X-Frame-Options", "DENY")
		c.Header("X-XSS-Protection", "1; mode=block")
		c.Header("Strict-Transport-Security", "max-age=31536000; includeSubDomains")

		path := c.Request.URL.Path

		if path == "/health/live" ||
			path == "/health/ready" ||
			path == "/metrics" ||
			strings.HasPrefix(path, "/playground") {
			c.Next()
			return
		}

		key := c.GetHeader("X-API-KEY")
		if key == "" || key != apiKey {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized access"})
			return
		}

		c.Next()
	}
}
