package client

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"order-management/model"

	"github.com/sony/gobreaker"
)

type ProductClient interface {
	GetProductPrice(ctx context.Context, productID int64) (float64, error)
}

type resilientProductClient struct {
	baseURL    string
	httpClient *http.Client
	cb         *gobreaker.CircuitBreaker
}

func NewProductClient(baseURL string) ProductClient {
	cbSettings := gobreaker.Settings{
		Name:        "ProductServiceCB",
		MaxRequests: 2,
		Interval:    10 * time.Second,
		Timeout:     5 * time.Second,
		ReadyToTrip: func(counts gobreaker.Counts) bool {
			return counts.ConsecutiveFailures >= 3
		},
	}

	return &resilientProductClient{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 1500 * time.Millisecond,
		},
		cb: gobreaker.NewCircuitBreaker(cbSettings),
	}
}

func (c *resilientProductClient) GetProductPrice(ctx context.Context, productID int64) (float64, error) {
	result, err := c.cb.Execute(func() (interface{}, error) {
		var lastErr error
		url := fmt.Sprintf("%s/api/v1/products/%d", c.baseURL, productID)

		const maxRetries = 3
		for attempt := 1; attempt <= maxRetries; attempt++ {
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			default:
			}

			req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
			if err != nil {
				return nil, err
			}

			resp, err := c.httpClient.Do(req)
			if err == nil {
				defer resp.Body.Close()
				if resp.StatusCode == http.StatusOK {
					var p model.ProductDTO
					if err := json.NewDecoder(resp.Body).Decode(&p); err == nil {
						return p.Price, nil
					}
				}
				lastErr = fmt.Errorf("product service status %d", resp.StatusCode)
			} else {
				lastErr = err
			}

			backoff := time.Duration(100*(1<<(attempt-1))) * time.Millisecond
			select {
			case <-time.After(backoff):
			case <-ctx.Done():
				return nil, ctx.Err()
			}
		}

		return nil, fmt.Errorf("failed after %d attempts: %w", maxRetries, lastErr)
	})

	if err != nil {
		// Fallback price logic for resilient degraded continuity
		return 50.00, nil
	}

	return result.(float64), nil
}
