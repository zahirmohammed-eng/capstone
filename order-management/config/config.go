package config

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	GatewayPort        string
	OrderServiceURL    string
	ProductServiceURL  string
	OrderServicePort   string
	ProductServicePort string

	// Order DB Config
	OrderDBHost     string
	OrderDBPort     string
	OrderDBName     string
	OrderDBUser     string
	OrderDBPassword string
	OrderDBSSLMode  string

	// Product DB Config
	ProductDBHost     string
	ProductDBPort     string
	ProductDBName     string
	ProductDBUser     string
	ProductDBPassword string
	ProductDBSSLMode  string

	APIKey string
}

func Load() (*Config, error) {
	_ = godotenv.Load()

	return &Config{
		GatewayPort:        getEnv("GATEWAY_PORT", "8080"),
		OrderServiceURL:    getEnv("ORDER_SERVICE_URL", "http://localhost:8081"),
		ProductServiceURL:  getEnv("PRODUCT_SERVICE_URL", "http://localhost:8082"),
		OrderServicePort:   getEnv("ORDER_SERVICE_PORT", "8081"),
		ProductServicePort: getEnv("PRODUCT_SERVICE_PORT", "8082"),

		OrderDBHost:     getEnv("ORDER_DB_HOST", "localhost"),
		OrderDBPort:     getEnv("ORDER_DB_PORT", "5432"),
		OrderDBName:     getEnv("ORDER_DB_NAME", "orderdb"),
		OrderDBUser:     getEnv("ORDER_DB_USER", "postgres"),
		OrderDBPassword: getEnv("ORDER_DB_PASSWORD", "postgres"),
		OrderDBSSLMode:  getEnv("ORDER_DB_SSLMODE", "disable"),

		ProductDBHost:     getEnv("PRODUCT_DB_HOST", "localhost"),
		ProductDBPort:     getEnv("PRODUCT_DB_PORT", "5433"),
		ProductDBName:     getEnv("PRODUCT_DB_NAME", "productdb"),
		ProductDBUser:     getEnv("PRODUCT_DB_USER", "postgres"),
		ProductDBPassword: getEnv("PRODUCT_DB_PASSWORD", "postgres"),
		ProductDBSSLMode:  getEnv("PRODUCT_DB_SSLMODE", "disable"),

		APIKey: getEnv("API_KEY", "secret-api-key"),
	}, nil
}

func (c *Config) OrderDSN() string {
	return fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		c.OrderDBHost, c.OrderDBPort, c.OrderDBUser, c.OrderDBPassword, c.OrderDBName, c.OrderDBSSLMode)
}

func (c *Config) ProductDSN() string {
	return fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		c.ProductDBHost, c.ProductDBPort, c.ProductDBUser, c.ProductDBPassword, c.ProductDBName, c.ProductDBSSLMode)
}

func getEnv(key, fallback string) string {
	if val, ok := os.LookupEnv(key); ok && val != "" {
		return val
	}
	return fallback
}
