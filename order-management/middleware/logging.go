package middleware

import (
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
)

func StructuredLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		c.Next()

		duration := time.Since(start)
		reqID := c.GetString(RequestIDHeader)
		status := c.Writer.Status()
		errMsg := c.Errors.ByType(gin.ErrorTypePrivate).String()

		fmt.Printf("request_id=%s method=%s path=%s status=%d duration=%s error=%q\n",
			reqID, c.Request.Method, path, status, duration, errMsg)
	}
}
