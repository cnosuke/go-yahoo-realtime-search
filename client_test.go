package yrs

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"
)

func newTestServer(fixturePath string) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		data, err := os.ReadFile(fixturePath)
		if err != nil {
			http.Error(w, "fixture not found", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=UTF-8")
		w.Write(data)
	}))
}

func TestSearch_Normal(t *testing.T) {
	ts := newTestServer("testdata/normal.html")
	defer ts.Close()

	client := NewClient(WithHTTPClient(ts.Client()), WithRequestTimeout(5*time.Second))
	client.scraper.baseURL = ts.URL

	result, err := client.Search(context.Background(), "golang")
	if err != nil {
		t.Fatalf("Search failed: %v", err)
	}
	if len(result.Tweets) != 3 {
		t.Fatalf("expected 3 tweets, got %d", len(result.Tweets))
	}
	if result.Query != "golang" {
		t.Errorf("expected query=golang, got %s", result.Query)
	}
	if result.Tweets[0].ScreenName != "testuser" {
		t.Errorf("expected screenName=testuser, got %s", result.Tweets[0].ScreenName)
	}
	if result.Tweets[0].Text != "Hello golang world!" {
		t.Errorf("expected normalized text, got %q", result.Tweets[0].Text)
	}
}

func TestSearch_NoMatch(t *testing.T) {
	ts := newTestServer("testdata/nomatch.html")
	defer ts.Close()

	client := NewClient(WithHTTPClient(ts.Client()), WithRequestTimeout(5*time.Second))
	client.scraper.baseURL = ts.URL

	result, err := client.Search(context.Background(), "nonexistent")
	if err != nil {
		t.Fatalf("Search failed: %v", err)
	}
	if len(result.Tweets) != 0 {
		t.Errorf("expected 0 tweets for no-match, got %d", len(result.Tweets))
	}
}

func TestSearch_EmptyQuery(t *testing.T) {
	client := NewClient()
	_, err := client.Search(context.Background(), "")
	if err == nil {
		t.Fatal("expected error for empty query")
	}
	if !errors.Is(err, ErrInvalidParameter) {
		t.Errorf("expected ErrInvalidParameter, got: %v", err)
	}
}

func TestSearchWithLimit_NegativeLimit(t *testing.T) {
	client := NewClient()
	_, err := client.SearchWithLimit(context.Background(), "test", -1)
	if err == nil {
		t.Fatal("expected error for negative limit")
	}
	if !errors.Is(err, ErrInvalidParameter) {
		t.Errorf("expected ErrInvalidParameter, got: %v", err)
	}
}

func TestSearch_WithLimit(t *testing.T) {
	ts := newTestServer("testdata/normal.html")
	defer ts.Close()

	client := NewClient(WithHTTPClient(ts.Client()), WithRequestTimeout(5*time.Second))
	client.scraper.baseURL = ts.URL

	result, err := client.SearchWithLimit(context.Background(), "golang", 2)
	if err != nil {
		t.Fatalf("SearchWithLimit failed: %v", err)
	}
	if len(result.Tweets) != 2 {
		t.Fatalf("expected 2 tweets with limit, got %d", len(result.Tweets))
	}
}

func TestSearch_EmptyEntriesWithoutNomatch(t *testing.T) {
	ts := newTestServer("testdata/empty_entries.html")
	defer ts.Close()

	client := NewClient(WithHTTPClient(ts.Client()), WithRequestTimeout(5*time.Second))
	client.scraper.baseURL = ts.URL

	_, err := client.Search(context.Background(), "test")
	if err == nil {
		t.Fatal("expected error for empty entries without #nomatch")
	}
	if !errors.Is(err, ErrScrapeFailed) {
		t.Errorf("expected ErrScrapeFailed, got: %v", err)
	}
}

func TestSearch_Non200Status(t *testing.T) {
	statuses := []int{http.StatusForbidden, http.StatusTooManyRequests, http.StatusServiceUnavailable}

	for _, statusCode := range statuses {
		code := statusCode
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(code)
		}))

		s := &scraper{httpClient: ts.Client(), userAgent: DefaultUserAgent, baseURL: ts.URL}
		_, err := s.fetch(context.Background(), "test", 0)
		ts.Close()

		if err == nil {
			t.Errorf("expected error for status %d", code)
			continue
		}
		if !errors.Is(err, ErrScrapeFailed) {
			t.Errorf("expected ErrScrapeFailed for status %d, got: %v", code, err)
		}
	}
}

func TestNewClient_Options(t *testing.T) {
	customClient := &http.Client{Timeout: 10 * time.Second}
	client := NewClient(
		WithHTTPClient(customClient),
		WithUserAgent("custom-agent/1.0"),
		WithRequestTimeout(15*time.Second),
	)

	if client.config.UserAgent != "custom-agent/1.0" {
		t.Errorf("expected custom user agent, got %s", client.config.UserAgent)
	}
	if client.config.RequestTimeout != 15*time.Second {
		t.Errorf("expected 15s timeout, got %v", client.config.RequestTimeout)
	}
	if client.scraper.userAgent != "custom-agent/1.0" {
		t.Errorf("expected scraper user agent=custom-agent/1.0, got %s", client.scraper.userAgent)
	}
}

func TestSearch_ContextCancelled(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(5 * time.Second)
	}))
	defer ts.Close()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	client := NewClient(WithHTTPClient(ts.Client()), WithRequestTimeout(0))
	client.scraper.baseURL = ts.URL
	_, err := client.Search(ctx, "test")
	if err == nil {
		t.Fatal("expected error for cancelled context")
	}
}
