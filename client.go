/*
Package yrs provides a Go client library for searching tweets (X/Twitter posts)
via Yahoo! Japan Realtime Search. It scrapes the __NEXT_DATA__ JSON from Yahoo!
Realtime Search pages, providing a clean, typed API without requiring any API
keys or authentication.

Basic Usage:

	client := yrs.NewClient()
	result, err := client.Search(context.Background(), "golang")
	if err != nil {
	    log.Fatal(err)
	}
	for _, tw := range result.Tweets {
	    fmt.Printf("@%s: %s\n", tw.ScreenName, tw.Text)
	}
*/
package yrs

import (
	"context"

	ierrors "github.com/cnosuke/go-yahoo-realtime-search/internal/errors"
)

// Client is the Yahoo Realtime Search client.
type Client struct {
	config  ClientConfig
	scraper *scraper
}

// NewClient creates a new Yahoo Realtime Search client.
func NewClient(opts ...ClientOption) *Client {
	cfg := newDefaultClientConfig()
	for _, opt := range opts {
		opt(cfg)
	}

	return &Client{
		config: *cfg,
		scraper: &scraper{
			httpClient: cfg.HTTPClient,
			userAgent:  cfg.UserAgent,
			baseURL:    baseURL,
		},
	}
}

// Search performs a keyword search and returns tweets.
func (c *Client) Search(ctx context.Context, query string) (*SearchResult, error) {
	return c.SearchWithLimit(ctx, query, 0)
}

// SearchWithLimit performs a search with a maximum number of results.
// A limit of 0 means no limit (return all results from the page).
func (c *Client) SearchWithLimit(ctx context.Context, query string, limit int) (*SearchResult, error) {
	if query == "" {
		return nil, ierrors.Wrap(ErrInvalidParameter, "query cannot be empty")
	}
	if limit < 0 {
		return nil, ierrors.Wrapf(ErrInvalidParameter, "limit must be non-negative, got %d", limit)
	}

	var cancelFunc context.CancelFunc = func() {}
	if c.config.RequestTimeout > 0 {
		if _, ok := ctx.Deadline(); !ok {
			ctx, cancelFunc = context.WithTimeout(ctx, c.config.RequestTimeout)
		}
	}
	defer cancelFunc()

	return c.scraper.fetch(ctx, query, limit)
}
