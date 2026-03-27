package yrs

import (
	"os"
	"strings"
	"testing"
	"time"
)

func TestNormalizeText(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"Hello \tSTART\tgolang\tEND\t world!", "Hello golang world!"},
		{"No markers here", "No markers here"},
		{"\tSTART\ta\tEND\t and \tSTART\tb\tEND\t", "a and b"},
		{"", ""},
	}

	for _, tt := range tests {
		got := normalizeText(tt.input)
		if got != tt.want {
			t.Errorf("normalizeText(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestParseNextData_Normal(t *testing.T) {
	f, err := os.Open("testdata/normal.html")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	s := &scraper{}
	nd, noMatch, err := s.parseNextData(f)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if noMatch {
		t.Fatal("expected noMatch=false")
	}

	entries := nd.Props.PageProps.PageData.Timeline.Entry
	if len(entries) != 3 {
		t.Fatalf("expected 3 entries, got %d", len(entries))
	}

	if entries[0].ID != "1234567890" {
		t.Errorf("expected first entry ID=1234567890, got %s", entries[0].ID)
	}
	if entries[0].ScreenName != "testuser" {
		t.Errorf("expected screenName=testuser, got %s", entries[0].ScreenName)
	}
}

func TestParseNextData_NoMatch(t *testing.T) {
	f, err := os.Open("testdata/nomatch.html")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	s := &scraper{}
	_, noMatch, err := s.parseNextData(f)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !noMatch {
		t.Fatal("expected noMatch=true")
	}
}

func TestParseNextData_MissingNextData(t *testing.T) {
	f, err := os.Open("testdata/no_next_data.html")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	s := &scraper{}
	_, _, err = s.parseNextData(f)
	if err == nil {
		t.Fatal("expected error for missing __NEXT_DATA__")
	}
	if !strings.Contains(err.Error(), "__NEXT_DATA__") {
		t.Errorf("error should mention __NEXT_DATA__, got: %v", err)
	}
}

func TestParseNextData_MalformedJSON(t *testing.T) {
	f, err := os.Open("testdata/malformed_json.html")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	s := &scraper{}
	_, _, err = s.parseNextData(f)
	if err == nil {
		t.Fatal("expected error for malformed JSON")
	}
	if !strings.Contains(err.Error(), "unmarshal") {
		t.Errorf("error should mention unmarshal, got: %v", err)
	}
}

func TestToTweets(t *testing.T) {
	nd := &nextData{}
	nd.Props.PageProps.PageData.Timeline.Entry = []timelineEntry{
		{
			ID:         "111",
			Text:       "Hello \tSTART\tworld\tEND\t",
			URL:        "https://x.com/user1/status/111",
			ScreenName: "user1",
			Name:       "User One",
			CreatedAt:  1711500000,
			ReplyCount: 1,
			RTCount:    2,
			LikeCount:  3,
			Media: []mediaEntry{
				{Type: "image", Item: mediaItem{MediaURL: "https://example.com/img.jpg"}},
			},
		},
		{
			ID:         "222",
			Text:       "Plain text",
			URL:        "",
			ScreenName: "user2",
			Name:       "User Two",
			CreatedAt:  1711500100,
		},
	}

	tweets := toTweets(nd, 0)
	if len(tweets) != 2 {
		t.Fatalf("expected 2 tweets, got %d", len(tweets))
	}

	if tweets[0].Text != "Hello world" {
		t.Errorf("expected normalized text, got %q", tweets[0].Text)
	}
	if len(tweets[0].Images) != 1 {
		t.Errorf("expected 1 image, got %d", len(tweets[0].Images))
	}
	if tweets[0].CreatedAt != time.Unix(1711500000, 0) {
		t.Errorf("unexpected CreatedAt: %v", tweets[0].CreatedAt)
	}

	// URL fallback when empty
	if tweets[1].URL != "https://x.com/user2/status/222" {
		t.Errorf("expected fallback URL, got %q", tweets[1].URL)
	}
}

func TestToTweets_WithLimit(t *testing.T) {
	nd := &nextData{}
	nd.Props.PageProps.PageData.Timeline.Entry = []timelineEntry{
		{ID: "1", Text: "a", ScreenName: "u1", Name: "U1", CreatedAt: 1},
		{ID: "2", Text: "b", ScreenName: "u2", Name: "U2", CreatedAt: 2},
		{ID: "3", Text: "c", ScreenName: "u3", Name: "U3", CreatedAt: 3},
	}

	tweets := toTweets(nd, 2)
	if len(tweets) != 2 {
		t.Fatalf("expected 2 tweets with limit, got %d", len(tweets))
	}
	if tweets[0].ID != "1" || tweets[1].ID != "2" {
		t.Error("expected first two entries")
	}
}
