package main

import (
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"

	"order-management/config"
	"order-management/middleware"

	"github.com/gin-gonic/gin"
)

func createReverseProxy(target string) gin.HandlerFunc {
	targetURL, err := url.Parse(target)
	if err != nil {
		log.Fatalf("Fatal: Invalid proxy target %s: %v", target, err)
	}

	proxy := httputil.NewSingleHostReverseProxy(targetURL)

	originalDirector := proxy.Director
	proxy.Director = func(req *http.Request) {
		originalDirector(req)
		req.Header.Set("X-Forwarded-Host", req.Header.Get("Host"))
		req.Host = targetURL.Host
	}

	return func(c *gin.Context) {
		proxy.ServeHTTP(c.Writer, c.Request)
	}
}

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Gateway failed to load configuration: %v", err)
	}

	r := gin.New()
	r.Use(
		middleware.Recovery(),
		middleware.RequestID(),
		middleware.StructuredLogger(),
		middleware.MetricsCollector(),
		middleware.SecurityAndAuth(cfg.APIKey),
	)

	orderProxy := createReverseProxy(cfg.OrderServiceURL)
	productProxy := createReverseProxy(cfg.ProductServiceURL)

	// Route mapping to Order Service (:8081)
	r.Any("/api/v1/orders", orderProxy)
	r.Any("/api/v1/orders/*path", orderProxy)
	r.Any("/graphql", orderProxy)
	r.Any("/playground", orderProxy)

	// Route mapping to Product Service (:8082)
	r.Any("/api/v1/products", productProxy)
	r.Any("/api/v1/products/*path", productProxy)

	r.GET("/health/live", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "GATEWAY_ALIVE"})
	})

	r.GET("/metrics", func(c *gin.Context) {
		c.JSON(http.StatusOK, middleware.GlobalMetrics)
	})

	log.Printf(" API Gateway online on public port :%s", cfg.GatewayPort)
	if err := r.Run(":" + cfg.GatewayPort); err != nil {
		log.Fatalf("Gateway failure: %v", err)
	}
}
