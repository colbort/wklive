package itickrest

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"golang.org/x/time/rate"
)

// Client is the single exit point for all iTick REST endpoints.
type Client struct {
	token   string
	http    *http.Client
	limiter *rate.Limiter
}

func New(token string, limiter *rate.Limiter, httpClient *http.Client) *Client {
	if httpClient == nil {
		httpClient = &http.Client{Timeout: 20 * time.Second}
	}
	return &Client{token: strings.TrimSpace(token), http: httpClient, limiter: limiter}
}

func (c *Client) Get(ctx context.Context, rawURL string) (*http.Response, error) {
	if c == nil {
		return nil, fmt.Errorf("itick REST client is not configured")
	}
	if c.limiter != nil {
		if err := c.limiter.Wait(ctx); err != nil {
			return nil, err
		}
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("accept", "application/json")
	req.Header.Set("token", c.token)
	resp, err := c.http.Do(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode >= http.StatusOK && resp.StatusCode < http.StatusMultipleChoices {
		return resp, nil
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(io.LimitReader(resp.Body, 2048))
	return nil, fmt.Errorf("itick REST returned status=%d body=%s", resp.StatusCode, strings.TrimSpace(string(body)))
}
