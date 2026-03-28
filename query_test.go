package yrs

import (
	"errors"
	"testing"
)

func TestQueryBuild(t *testing.T) {
	tests := []struct {
		name    string
		query   *Query
		want    string
		wantErr bool
	}{
		{
			name:  "single keyword",
			query: NewQuery("golang"),
			want:  "golang",
		},
		{
			name:  "multiple AND keywords",
			query: NewQuery("golang", "tutorial"),
			want:  "golang tutorial",
		},
		{
			name:  "And chaining",
			query: NewQuery("golang").And("tutorial"),
			want:  "golang tutorial",
		},
		{
			name:  "OR group",
			query: NewQuery("猫").Or("犬", "うさぎ"),
			want:  "猫 (犬 うさぎ)",
		},
		{
			name:  "NOT keyword",
			query: NewQuery("golang").Not("tutorial"),
			want:  "golang -tutorial",
		},
		{
			name:  "FromUser",
			query: NewQuery("news").FromUser("YahooSearchJP"),
			want:  "news ID:YahooSearchJP",
		},
		{
			name:  "ToUser only",
			query: NewQuery().ToUser("YahooSearchJP"),
			want:  "@YahooSearchJP",
		},
		{
			name:  "Hashtag only",
			query: NewQuery().Hashtag("地震"),
			want:  "#地震",
		},
		{
			name:  "URL filter only",
			query: NewQuery().URL("yahoo.co.jp"),
			want:  "URL:yahoo.co.jp",
		},
		{
			name:  "complex query",
			query: NewQuery("golang").Not("tutorial").FromUser("user1").Hashtag("go"),
			want:  "golang -tutorial ID:user1 #go",
		},
		{
			name:  "only OR group",
			query: NewQuery().Or("猫", "犬"),
			want:  "(猫 犬)",
		},
		{
			name:  "multiple OR groups",
			query: NewQuery().Or("猫", "犬").Or("赤", "青"),
			want:  "(猫 犬) (赤 青)",
		},
		{
			name:    "empty query",
			query:   NewQuery(),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.query.Build()
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				if !errors.Is(err, ErrInvalidParameter) {
					t.Fatalf("expected ErrInvalidParameter, got %v", err)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.want {
				t.Errorf("got %q, want %q", got, tt.want)
			}
		})
	}
}

func TestQueryValidation(t *testing.T) {
	tests := []struct {
		name  string
		query *Query
	}{
		{
			name:  "full-width hash in hashtag",
			query: NewQuery().Hashtag("＃地震"),
		},
		{
			name:  "full-width at in ToUser",
			query: NewQuery().ToUser("＠user"),
		},
		{
			name:  "full-width minus in NOT",
			query: NewQuery("test").Not("－bad"),
		},
		{
			name:  "empty query via no-op FromUser",
			query: NewQuery().FromUser(""),
		},
		{
			name:  "empty keyword",
			query: NewQuery(""),
		},
		{
			name:  "empty OR keyword",
			query: NewQuery("test").Or("", "b"),
		},
		{
			name:  "empty NOT keyword",
			query: NewQuery("test").Not(""),
		},
		{
			name:  "URL with URL: prefix",
			query: NewQuery("test").URL("URL:example.com"),
		},
		{
			name:  "hashtag with space",
			query: NewQuery().Hashtag("hello world"),
		},
		{
			name:  "hashtag with # prefix",
			query: NewQuery().Hashtag("#地震"),
		},
		{
			name:  "FromUser with ID: prefix",
			query: NewQuery().FromUser("ID:user"),
		},
		{
			name:  "ToUser with @ prefix",
			query: NewQuery().ToUser("@user"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := tt.query.Build()
			if err == nil {
				t.Fatal("expected error, got nil")
			}
			if !errors.Is(err, ErrInvalidParameter) {
				t.Fatalf("expected ErrInvalidParameter, got %v", err)
			}
		})
	}
}
