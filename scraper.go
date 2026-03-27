package yrs

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"time"

	"github.com/PuerkitoBio/goquery"
	ierrors "github.com/cnosuke/go-yahoo-realtime-search/internal/errors"
)

type scraper struct {
	httpClient *http.Client
	userAgent  string
	baseURL    string
}

type nextData struct {
	Props struct {
		PageProps struct {
			PageData struct {
				Timeline timeline `json:"timeline"`
			} `json:"pageData"`
		} `json:"pageProps"`
	} `json:"props"`
}

type timeline struct {
	Entry []timelineEntry `json:"entry"`
}

type mediaItem struct {
	MediaURL string `json:"mediaUrl"`
}

type mediaEntry struct {
	Type string `json:"type"`
	Item mediaItem `json:"item"`
}

type timelineEntry struct {
	ID         string `json:"id"`
	Text       string `json:"displayText"`
	URL        string `json:"url"`
	ScreenName string `json:"screenName"`
	Name       string `json:"name"`
	CreatedAt  int64  `json:"createdAt"`
	ReplyCount int    `json:"replyCount"`
	RTCount    int    `json:"rtCount"`
	LikeCount  int    `json:"likesCount"`
	Media []mediaEntry `json:"media"`
}

var enclosureRegexp = regexp.MustCompile(`\tSTART\t(.*?)\tEND\t`)

func normalizeText(text string) string {
	return enclosureRegexp.ReplaceAllString(text, "$1")
}

func (s *scraper) fetch(ctx context.Context, query string, limit int) (*SearchResult, error) {
	u, err := url.Parse(s.baseURL)
	if err != nil {
		return nil, ierrors.Wrapf(ErrScrapeFailed, "failed to parse base URL: %v", err)
	}
	q := u.Query()
	q.Set("p", query)
	q.Set("ei", "UTF-8")
	u.RawQuery = q.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, ierrors.Wrapf(ErrScrapeFailed, "failed to create request: %v", err)
	}
	req.Header.Set("User-Agent", s.userAgent)

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, ierrors.Wrapf(ErrScrapeFailed, "HTTP request failed for %s: %v", u.String(), err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		io.Copy(io.Discard, io.LimitReader(resp.Body, 4096))
		return nil, ierrors.Wrapf(ErrScrapeFailed, "unexpected HTTP status %d for URL %s", resp.StatusCode, u.String())
	}

	nd, noMatch, err := s.parseNextData(io.LimitReader(resp.Body, 10*1024*1024))
	if err != nil {
		return nil, err
	}

	if noMatch {
		return &SearchResult{Query: query, Tweets: nil}, nil
	}

	entries := nd.Props.PageProps.PageData.Timeline.Entry
	if len(entries) == 0 {
		return nil, ierrors.Wrap(ErrScrapeFailed, "no entries found despite non-nomatch page; page structure may have changed")
	}

	tweets := toTweets(nd, limit)
	return &SearchResult{Query: query, Tweets: tweets}, nil
}

func (s *scraper) parseNextData(htmlBody io.Reader) (*nextData, bool, error) {
	doc, err := goquery.NewDocumentFromReader(htmlBody)
	if err != nil {
		return nil, false, ierrors.Wrapf(ErrScrapeFailed, "failed to parse HTML: %v", err)
	}

	if doc.Find("#nomatch").Length() > 0 {
		return nil, true, nil
	}

	scriptTag := doc.Find("#__NEXT_DATA__")
	if scriptTag.Length() == 0 {
		return nil, false, ierrors.Wrap(ErrScrapeFailed, "__NEXT_DATA__ script tag not found")
	}

	jsonStr := scriptTag.Text()
	var nd nextData
	if err := json.Unmarshal([]byte(jsonStr), &nd); err != nil {
		return nil, false, ierrors.Wrapf(ErrScrapeFailed, "failed to unmarshal __NEXT_DATA__ JSON: %v", err)
	}

	return &nd, false, nil
}

func toTweets(nd *nextData, limit int) []Tweet {
	entries := nd.Props.PageProps.PageData.Timeline.Entry
	if limit > 0 && limit < len(entries) {
		entries = entries[:limit]
	}

	tweets := make([]Tweet, 0, len(entries))
	for _, e := range entries {
		var images []Image
		for _, m := range e.Media {
			if m.Type == "image" && m.Item.MediaURL != "" {
				images = append(images, Image{URL: m.Item.MediaURL})
			}
		}

		tweetURL := e.URL
		if tweetURL == "" && e.ScreenName != "" && e.ID != "" {
			tweetURL = fmt.Sprintf("https://x.com/%s/status/%s", e.ScreenName, e.ID)
		}

		tweets = append(tweets, Tweet{
			ID:         e.ID,
			URL:        tweetURL,
			Text:       normalizeText(e.Text),
			AuthorName: e.Name,
			ScreenName: e.ScreenName,
			CreatedAt:  time.Unix(e.CreatedAt, 0),
			ReplyCount: e.ReplyCount,
			RTCount:    e.RTCount,
			LikeCount:  e.LikeCount,
			Images:     images,
		})
	}
	return tweets
}
