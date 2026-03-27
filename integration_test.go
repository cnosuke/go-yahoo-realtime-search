//go:build integration

package yrs

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestIntegration_SearchCommonQuery(t *testing.T) {
	client := NewClient(WithRequestTimeout(30 * time.Second))
	result, err := client.Search(context.Background(), "Twitter")
	if err != nil {
		t.Fatalf("Search failed: %v", err)
	}
	if len(result.Tweets) == 0 {
		t.Fatal("expected at least one tweet for query 'Twitter'")
	}
	if result.Query != "Twitter" {
		t.Errorf("expected query=Twitter, got %s", result.Query)
	}

	tw := result.Tweets[0]
	if tw.ID == "" {
		t.Error("expected non-empty tweet ID")
	}
	if tw.Text == "" {
		t.Error("expected non-empty tweet text")
	}
	if tw.ScreenName == "" {
		t.Error("expected non-empty screen name")
	}
}

func TestIntegration_SearchNoResults(t *testing.T) {
	client := NewClient(WithRequestTimeout(30 * time.Second))
	randomQuery := uuid.New().String()
	result, err := client.Search(context.Background(), randomQuery)
	if err != nil {
		t.Fatalf("Search failed: %v", err)
	}
	if len(result.Tweets) != 0 {
		t.Errorf("expected 0 tweets for random query, got %d", len(result.Tweets))
	}
}

func TestIntegration_SearchWithLimit(t *testing.T) {
	client := NewClient(WithRequestTimeout(30 * time.Second))
	result, err := client.SearchWithLimit(context.Background(), "Twitter", 3)
	if err != nil {
		t.Fatalf("SearchWithLimit failed: %v", err)
	}
	if len(result.Tweets) > 3 {
		t.Errorf("expected at most 3 tweets, got %d", len(result.Tweets))
	}
}
