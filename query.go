package yrs

import (
	"strings"

	ierrors "github.com/cnosuke/go-yahoo-realtime-search/internal/errors"
)

// Query builds a Yahoo Realtime Search query string.
type Query struct {
	keywords  []string
	orGroups  [][]string
	nots      []string
	fromUser  string
	toUser    string
	hashtags  []string
	urlFilter string
}

// NewQuery creates a new Query with the given AND keywords.
func NewQuery(keywords ...string) *Query {
	return &Query{keywords: keywords}
}

// And adds additional AND keywords.
func (q *Query) And(keywords ...string) *Query {
	q.keywords = append(q.keywords, keywords...)
	return q
}

// Or adds an OR group. Keywords within the group are OR-ed together.
func (q *Query) Or(keywords ...string) *Query {
	if len(keywords) > 0 {
		q.orGroups = append(q.orGroups, keywords)
	}
	return q
}

// Not excludes posts containing the given keywords.
func (q *Query) Not(keywords ...string) *Query {
	q.nots = append(q.nots, keywords...)
	return q
}

// FromUser filters posts from a specific account (ID:username).
func (q *Query) FromUser(username string) *Query {
	q.fromUser = username
	return q
}

// ToUser filters posts mentioning/replying to a specific account (@username).
func (q *Query) ToUser(username string) *Query {
	q.toUser = username
	return q
}

// Hashtag filters posts containing specific hashtags.
func (q *Query) Hashtag(tags ...string) *Query {
	q.hashtags = append(q.hashtags, tags...)
	return q
}

// URL filters posts containing a specific URL or domain (URL:xxx, prefix match).
func (q *Query) URL(urlOrDomain string) *Query {
	q.urlFilter = urlOrDomain
	return q
}

// Build generates the query string. Returns an error if validation fails.
func (q *Query) Build() (string, error) {
	if err := q.validate(); err != nil {
		return "", err
	}

	var parts []string

	parts = append(parts, q.keywords...)

	for _, group := range q.orGroups {
		parts = append(parts, "("+strings.Join(group, " ")+")")
	}

	for _, kw := range q.nots {
		parts = append(parts, "-"+kw)
	}

	if q.fromUser != "" {
		parts = append(parts, "ID:"+q.fromUser)
	}

	if q.toUser != "" {
		parts = append(parts, "@"+q.toUser)
	}

	for _, tag := range q.hashtags {
		parts = append(parts, "#"+tag)
	}

	if q.urlFilter != "" {
		parts = append(parts, "URL:"+q.urlFilter)
	}

	result := strings.Join(parts, " ")
	if result == "" {
		return "", ierrors.Wrap(ErrInvalidParameter, "query cannot be empty")
	}

	return result, nil
}

func (q *Query) validate() error {
	for _, kw := range q.keywords {
		if err := validateNotEmpty("keyword", kw); err != nil {
			return err
		}
		if err := validateNoFullWidthSymbols(kw); err != nil {
			return err
		}
	}
	for _, group := range q.orGroups {
		for _, kw := range group {
			if err := validateNotEmpty("OR keyword", kw); err != nil {
				return err
			}
			if err := validateNoFullWidthSymbols(kw); err != nil {
				return err
			}
		}
	}
	for _, kw := range q.nots {
		if err := validateNotEmpty("NOT keyword", kw); err != nil {
			return err
		}
		if err := validateNoFullWidthSymbols(kw); err != nil {
			return err
		}
	}

	if q.fromUser != "" {
		// validateNotEmpty catches whitespace-only strings (e.g. "  ")
		if err := validateNotEmpty("fromUser", q.fromUser); err != nil {
			return err
		}
		if err := validateNoPrefix("fromUser", q.fromUser, "ID:"); err != nil {
			return err
		}
		if err := validateNoFullWidthSymbols(q.fromUser); err != nil {
			return err
		}
	}

	if q.toUser != "" {
		// validateNotEmpty catches whitespace-only strings (e.g. "  ")
		if err := validateNotEmpty("toUser", q.toUser); err != nil {
			return err
		}
		if err := validateNoPrefix("toUser", q.toUser, "@"); err != nil {
			return err
		}
		if err := validateNoFullWidthSymbols(q.toUser); err != nil {
			return err
		}
	}

	for _, tag := range q.hashtags {
		if err := validateNotEmpty("hashtag", tag); err != nil {
			return err
		}
		if err := validateNoPrefix("hashtag", tag, "#"); err != nil {
			return err
		}
		if strings.Contains(tag, " ") {
			return ierrors.Wrapf(ErrInvalidParameter, "hashtag must not contain spaces: %q", tag)
		}
		if err := validateNoFullWidthSymbols(tag); err != nil {
			return err
		}
	}

	if q.urlFilter != "" {
		if err := validateNoPrefix("URL", q.urlFilter, "URL:"); err != nil {
			return err
		}
		if err := validateNoFullWidthSymbols(q.urlFilter); err != nil {
			return err
		}
	}

	return nil
}
