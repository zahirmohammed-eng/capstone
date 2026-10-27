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
	"order-management/services/product/controller"
	"order-management/services/product/repository"
	"order-management/services/product/service"

	"github.com/gin-gonic/gin"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Product service config failure: %v", err)
	}

	// Initialize independent GORM database for Product Service
	gormDB, err := database.InitGormDB(cfg.ProductDSN(), "product_service.db")
	if err != nil {
		log.Fatalf("Failed to initialize Product Database: %v", err)
	}

	// GORM AutoMigrate schema
	if err := gormDB.AutoMigrate(&model.Product{}); err != nil {
		log.Fatalf("Product AutoMigrate failed: %v", err)
	}

	// Layered Architecture Dependency Injection
	repo := repository.NewGormProductRepository(gormDB)
	_ = repo.SeedDefaults(context.Background()) // seed products 501, 502, 503
	svc := service.NewProductService(repo)
	ctrl := controller.NewProductController(svc)

	r := gin.New()
	r.Use(gin.Recovery())

	// Product Endpoints
	r.POST("/api/v1/products", ctrl.Create)
	r.GET("/api/v1/products/:id", ctrl.GetByID)
	r.GET("/api/v1/products", ctrl.List)

	r.GET("/health/live", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "PRODUCT_SERVICE_ALIVE"})
	})

	srv := &http.Server{
		Addr:    ":" + cfg.ProductServicePort,
		Handler: r,
	}

	go func() {
		log.Printf("Product Service running with GORM on port :%s", cfg.ProductServicePort)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Product Service crashed: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down Product server gracefully...")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_ = srv.Shutdown(ctx)
	log.Println("Product server exited cleanly.")
}
