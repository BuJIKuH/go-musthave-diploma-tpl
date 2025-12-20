package accrual

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"
)

const (
	StatusRegistered = "REGISTERED"
	StatusProcessing = "PROCESSING"
	StatusInvalid    = "INVALID"
	StatusProcessed  = "PROCESSED"
)

var (
	ErrTooManyRequests = errors.New("too many requests")
	ErrOrderNotFound   = errors.New("order not found")
)

type OrderResponse struct {
	Order   string `json:"order"`
	Status  string `json:"status"`
	Accrual string `json:"accrual,omitempty"`
}

type Client struct {
	baseURL string
	client  *http.Client
}

func NewClient(baseURL string) *Client {
	return &Client{
		baseURL: baseURL,
		client: &http.Client{
			Timeout: 5 * time.Second,
		},
	}
}

func (c *Client) GetOrder(ctx context.Context, number string) (*OrderResponse, error) {
	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodGet,
		fmt.Sprintf("%s/api/orders/%s", c.baseURL, number),
		nil,
	)
	if err != nil {
		return nil, err
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusOK:
		var res OrderResponse
		if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
			return nil, err
		}
		return &res, nil

	case http.StatusNoContent:
		return nil, ErrOrderNotFound

	case http.StatusTooManyRequests:
		return nil, ErrTooManyRequests

	default:
		return nil, fmt.Errorf("unexpected status %d", resp.StatusCode)
	}
}
