package yrs

import "time"

// SearchResult holds the result of a Yahoo Realtime Search query.
type SearchResult struct {
	Query  string  `json:"query"`
	Tweets []Tweet `json:"tweets"`
}

// Tweet represents a single tweet extracted from Yahoo Realtime Search.
type Tweet struct {
	ID         string    `json:"id"`
	URL        string    `json:"url"`
	Text       string    `json:"text"`
	AuthorName string    `json:"author_name"`
	ScreenName string    `json:"screen_name"`
	CreatedAt  time.Time `json:"created_at"`

	ReplyCount int `json:"reply_count"`
	RTCount    int `json:"rt_count"`
	LikeCount  int `json:"like_count"`

	Images []Image `json:"images,omitempty"`
}

// Image represents an image attached to a tweet.
type Image struct {
	URL string `json:"url"`
}
