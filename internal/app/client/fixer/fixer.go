package fixer

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"

	"github.com/iagafon/worker-service/internal/app/config/section"
	"github.com/iagafon/worker-service/internal/app/entity"
)

// Client — HTTP клиент для работы с Fixer API.
type Client struct {
	httpClient *http.Client
	apiKey     string
	baseURL    string
}

// Response — ответ от Fixer API.
type Response struct {
	Success   bool               `json:"success"`
	Timestamp int64              `json:"timestamp"`
	Base      string             `json:"base"`
	Date      string             `json:"date"`
	Rates     map[string]float64 `json:"rates"`
	Error     *Error             `json:"error,omitempty"`
}

// Error — ошибка от Fixer API.
type Error struct {
	Code int    `json:"code"`
	Type string `json:"type"`
	Info string `json:"info"`
}

// NewClient создаёт новый клиент Fixer API.
func NewClient(cfg section.ClientFixer) *Client {
	httpClient := &http.Client{
		Transport: otelhttp.NewTransport(http.DefaultTransport),
		Timeout:   10 * time.Second,
	}

	return &Client{
		httpClient: httpClient,
		apiKey:     cfg.ApiKey,
		baseURL:    cfg.BaseURL,
	}
}

// GetRates получает курсы валют относительно базовой валюты.
func (c *Client) GetRates(ctx context.Context, base string) (map[string]float64, error) {
	args := make(url.Values)
	args.Set("access_key", c.apiKey)
	args.Set("base", base)

	requestURL := fmt.Sprintf("%s/latest?%s", c.baseURL, args.Encode())

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, requestURL, http.NoBody)
	if err != nil {
		return nil, fmt.Errorf("http.NewRequestWithContext: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("c.httpClient.Do: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, entity.ErrFixerUnavailable
	}

	var fixerResponse Response
	if err := json.NewDecoder(resp.Body).Decode(&fixerResponse); err != nil {
		return nil, entity.ErrFixerInvalidResponse
	}

	if !fixerResponse.Success {
		return nil, c.mapFixerError(fixerResponse.Error)
	}

	return fixerResponse.Rates, nil
}

// mapFixerError преобразует ошибку Fixer API в типизированную ошибку.
func (c *Client) mapFixerError(fixerErr *Error) error {
	if fixerErr == nil {
		return nil
	}

	switch fixerErr.Code {
	case 101: // invalid_access_key
		return entity.ErrFixerInvalidApiKey
	case 104, 105: // rate limit
		return entity.ErrFixerRateLimitExceeded
	default:
		return fmt.Errorf("%w: [%d] %s - %s", entity.ErrFixerInvalidResponse,
			fixerErr.Code, fixerErr.Type, fixerErr.Info)
	}
}
