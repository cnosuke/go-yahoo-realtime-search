package yrs

import (
	"net/http"
	"time"
)

// ClientConfig holds the resolved configuration for the client.
type ClientConfig struct {
	HTTPClient     *http.Client
	UserAgent      string
	RequestTimeout time.Duration
}

func newDefaultClientConfig() *ClientConfig {
	return &ClientConfig{
		HTTPClient:     &http.Client{},
		UserAgent:      DefaultUserAgent,
		RequestTimeout: DefaultRequestTimeout,
	}
}
