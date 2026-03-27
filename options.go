package yrs

import (
	"net/http"
	"time"
)

// ClientOption configures the Client.
type ClientOption func(*ClientConfig)

// WithHTTPClient sets a custom HTTP client.
func WithHTTPClient(client *http.Client) ClientOption {
	return func(cfg *ClientConfig) {
		cfg.HTTPClient = client
	}
}

// WithUserAgent sets a custom User-Agent header.
func WithUserAgent(ua string) ClientOption {
	return func(cfg *ClientConfig) {
		cfg.UserAgent = ua
	}
}

// WithRequestTimeout sets the timeout for HTTP requests.
func WithRequestTimeout(timeout time.Duration) ClientOption {
	return func(cfg *ClientConfig) {
		cfg.RequestTimeout = timeout
	}
}
