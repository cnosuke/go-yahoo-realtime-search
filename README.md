# go-yahoo-realtime-search

Go client library for searching tweets (X/Twitter posts) via [Yahoo! Japan Realtime Search](https://search.yahoo.co.jp/realtime). Scrapes the `__NEXT_DATA__` JSON from Yahoo! Realtime Search pages — no API keys or authentication required.

> **Note:** This library relies on web scraping. Yahoo! may change their page structure at any time, which could break functionality.

## Installation

```bash
go get github.com/cnosuke/go-yahoo-realtime-search
```

## Library Usage

```go
package main

import (
	"context"
	"fmt"
	"log"

	yrs "github.com/cnosuke/go-yahoo-realtime-search"
)

func main() {
	client := yrs.NewClient()

	result, err := client.Search(context.Background(), "golang")
	if err != nil {
		log.Fatal(err)
	}

	for _, tw := range result.Tweets {
		fmt.Printf("@%s: %s\n", tw.ScreenName, tw.Text)
	}
}
```

### Options

```go
client := yrs.NewClient(
	yrs.WithRequestTimeout(10 * time.Second),
	yrs.WithUserAgent("my-app/1.0"),
	yrs.WithHTTPClient(customHTTPClient),
)

// Limit number of results
result, err := client.SearchWithLimit(ctx, "query", 5)
```

### Types

```go
type SearchResult struct {
	Query  string  `json:"query"`
	Tweets []Tweet `json:"tweets"`
}

type Tweet struct {
	ID         string    `json:"id"`
	URL        string    `json:"url"`
	Text       string    `json:"text"`
	AuthorName string    `json:"author_name"`
	ScreenName string    `json:"screen_name"`
	CreatedAt  time.Time `json:"created_at"`
	ReplyCount int       `json:"reply_count"`
	RTCount    int       `json:"rt_count"`
	LikeCount  int       `json:"like_count"`
	Images     []Image   `json:"images,omitempty"`
}
```

### Error Handling

```go
if errors.Is(err, yrs.ErrScrapeFailed) {
	// Scraping failed (HTTP error, parsing error, etc.)
}
if errors.Is(err, yrs.ErrInvalidParameter) {
	// Invalid parameter (empty query, negative limit)
}
```

## CLI

A command-line tool `yrs` is included.

### Build

```bash
make build
# Binary: bin/yrs
```

### Usage

```bash
# Basic search
bin/yrs "golang"

# Limit results
bin/yrs -l 5 "golang"

# JSON output
bin/yrs --json "golang"

# Custom timeout
bin/yrs -t 10s "golang"
```

## Testing

```bash
# Unit tests
go test ./...

# Integration tests (hits live Yahoo)
go test -tags integration ./...
```

## License

MIT
