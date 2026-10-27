package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"order-management/config"
	"order-management/database"
	"order-management/model"
	"order-management/services/order/client"
	"order-management/services/order/controller"
	gqlserver "order-management/services/order/graphql"
	"order-management/services/order/repository"
	"order-management/services/order/service"

	"github.com/gin-gonic/gin"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Order service config failure: %v", err)
	}

	// Initialize independent GORM database for Order Service
	gormDB, err := database.InitGormDB(cfg.OrderDSN(), "order_service.db")
	if err != nil {
		log.Fatalf("Failed to initialize Order Database: %v", err)
	}

	// GORM AutoMigrate schema
	if err := gormDB.AutoMigrate(&model.Order{}); err != nil {
		log.Fatalf("Order AutoMigrate failed: %v", err)
	}

	// Layered Architecture Dependency Injection
	orderRepo := repository.NewGormOrderRepository(gormDB)
	productClient := client.NewProductClient(cfg.ProductServiceURL)
	orderService := service.NewOrderService(orderRepo, productClient)
	orderController := controller.NewOrderController(orderService)

	gs, err := gqlserver.NewServer(orderService)
	if err != nil {
		log.Fatalf("GraphQL schema initialization failed: %v", err)
	}
	gqlHandler := gqlserver.NewHTTPHandler(gs)

	r := gin.New()
	r.Use(gin.Recovery())

	api := r.Group("/api/v1")
	{
		api.POST("/orders", orderController.Create)
		api.GET("/orders/:id", orderController.GetByID)
		api.GET("/orders", orderController.List)
		api.PUT("/orders/:id/status", orderController.UpdateStatus)
	}

	r.Any("/graphql", gin.WrapH(gqlHandler))
	r.Any("/playground", gin.WrapH(gqlHandler))

	r.GET("/health/live", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ORDER_SERVICE_ALIVE"})
	})

	srv := &http.Server{
		Addr:    ":" + cfg.OrderServicePort,
		Handler: r,
	}

	go func() {
		log.Printf(" Order Service running with GORM on port :%s", cfg.OrderServicePort)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Order server failed: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down Order server gracefully...")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_ = srv.Shutdown(ctx)
	log.Println("Order server exited cleanly.")
}
