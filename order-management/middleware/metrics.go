package middleware

import (
	"sync/atomic"
	"time"

	"github.com/gin-gonic/gin"
)

type MetricsTracker struct {
	RequestCount uint64 `json:"request_count"`
	ErrorCount   uint64 `json:"error_count"`
	DBErrors     uint64 `json:"database_errors"`
	TotalLatency uint64 `json:"total_latency_ms"`
}

var GlobalMetrics = &MetricsTracker{}

func (m *MetricsTracker) IncDBErrors() {
	atomic.AddUint64(&m.DBErrors, 1)
}

func MetricsCollector() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()

		latency := time.Since(start).Milliseconds()
		atomic.AddUint64(&GlobalMetrics.RequestCount, 1)
		atomic.AddUint64(&GlobalMetrics.TotalLatency, uint64(latency))

		if c.Writer.Status() >= 400 {
			atomic.AddUint64(&GlobalMetrics.ErrorCount, 1)
		}
	}
}
