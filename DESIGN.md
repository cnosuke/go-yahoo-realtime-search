# Query Builder Design for go-yahoo-realtime-search

## Overview

Add a type-safe query builder API to `go-yahoo-realtime-search` that supports Yahoo! Realtime Search's official search syntax: AND/OR/NOT operators and ID/mention/hashtag/URL filters.

The query builder generates a query string that is passed to the existing `Search(ctx, query)` method. No changes to the scraper or HTTP layer.

## Yahoo! Realtime Search Syntax Reference

Source: https://support.yahoo-net.jp/SccRealtime/s/article/H000011629

### Operators

| Operator | Syntax | Example | Description |
|----------|--------|---------|-------------|
| AND | `keyword1 keyword2` (space-separated) | `猫 犬` | Posts containing ALL keywords |
| OR | `(keyword1 keyword2)` (parentheses + space) | `(猫 犬)` | Posts containing ANY keyword |
| NOT | `-keyword` (half-width minus prefix) | `猫 -犬` | Exclude posts containing keyword |

### Filters

| Filter | Syntax | Example | Description |
|--------|--------|---------|-------------|
| Account | `ID:username` | `ID:YahooSearchJP` | Posts FROM a specific account |
| Mention | `@username` | `@YahooSearchJP` | Posts TO/mentioning a specific account |
| Hashtag | `#tag` | `#地震` | Posts containing a hashtag |
| URL | `URL:domain_or_url` | `URL:yahoo.co.jp` | Posts containing URL (prefix match) |

### Key Rules

- All symbols (`-`, `()`, `@`, `#`, `ID:`, `URL:`) MUST be half-width characters.
- AND search: space can be full-width or half-width (we normalize to half-width).
- NOT: minus must be directly attached to the keyword, preceded by a space.
- OR: keywords inside parentheses are separated by spaces.

## Query Builder API

### New File: `query.go`

```go
package yrs

// Query builds a Yahoo Realtime Search query string.
// Use NewQuery to create a Query, chain methods, then call Build() or pass to SearchWithQuery().
type Query struct {
    keywords []string   // AND keywords
    orGroups [][]string // each group becomes (kw1 kw2 ...)
    nots     []string   // -keyword
    fromUser string     // ID:username
    toUser   string     // @username
    hashtags []string   // #tag
    urlFilter string   // URL:domain
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
// Each call adds a separate OR group: (kw1 kw2 ...).
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
// The username must NOT include the "ID:" prefix; Build() adds it automatically.
// Passing "ID:user" results in a validation error to prevent double-prefixing ("ID:ID:user").
func (q *Query) FromUser(username string) *Query {
    q.fromUser = username
    return q
}

// ToUser filters posts mentioning/replying to a specific account (@username).
// The username must NOT include the "@" prefix; Build() adds it automatically.
// Passing "@user" results in a validation error to prevent double-prefixing ("@@user").
func (q *Query) ToUser(username string) *Query {
    q.toUser = username
    return q
}

// Hashtag filters posts containing specific hashtags.
// The tag value must NOT include the leading "#" prefix; Build() adds it automatically.
// Passing "#tag" results in a validation error to prevent double-prefixing ("##tag").
func (q *Query) Hashtag(tags ...string) *Query {
    q.hashtags = append(q.hashtags, tags...)
    return q
}

// URL filters posts containing a specific URL or domain (URL:xxx, prefix match).
func (q *Query) URL(urlOrDomain string) *Query {
    q.urlFilter = urlOrDomain
    return q
}

// Build generates the query string.
// Returns the query string and an error if validation fails.
func (q *Query) Build() (string, error)
```

### Build() Implementation

`Build()` assembles parts in this order, joined by spaces:

1. AND keywords: `kw1 kw2 kw3`
2. OR groups: `(kw1 kw2)` per group
3. NOT keywords: `-kw1 -kw2`
4. FromUser: `ID:username`
5. ToUser: `@username`
6. Hashtags: `#tag1 #tag2`
7. URL filter: `URL:domain`

Example:

```go
q, err := yrs.NewQuery("golang").
    Not("tutorial").
    FromUser("YahooSearchJP").
    Build()
// q == "golang -tutorial ID:YahooSearchJP"
```

```go
q, err := yrs.NewQuery("猫").
    Or("犬", "うさぎ").
    Not("ペットショップ").
    Build()
// q == "猫 (犬 うさぎ) -ペットショップ"
```

### Validation

Validation runs inside `Build()`. Returns `ErrInvalidParameter` on failure.

#### Full-Width Character Detection

Detect common full-width equivalents of syntax-significant characters and return an error with a clear message:

| Full-Width | Half-Width | Context |
|------------|------------|---------|
| `＃` (U+FF03) | `#` | Hashtag |
| `＠` (U+FF20) | `@` | Mention |
| `（` (U+FF08) | `(` | OR group (internal) |
| `）` (U+FF09) | `)` | OR group (internal) |
| `－` (U+FF0D) | `-` | NOT (internal) |

```go
// validateNoFullWidthSymbols checks a string for full-width equivalents
// of search syntax characters and returns an error if found.
func validateNoFullWidthSymbols(s string) error
```

Validation targets:
- All keyword strings (AND, OR, NOT keywords)
- Username in FromUser/ToUser
- Hashtag values
- URL filter value

#### Leading-Prefix Validation

Detect accidental double-prefixing when users include the prefix that `Build()` adds automatically:

- `Hashtag("tag")` is correct; `Hashtag("#tag")` returns an error (would produce `##tag`).
- `FromUser("user")` is correct; `FromUser("ID:user")` returns an error (would produce `ID:ID:user`).
- `ToUser("user")` is correct; `ToUser("@user")` returns an error (would produce `@@user`).

#### Other Validations

- `Build()` returns error if the query would be empty (no keywords, no filters).
- Username must not be empty when `FromUser`/`ToUser` is called.
- Hashtag values must not contain spaces.

### New File: `validate.go`

```go
package yrs

import "strings"

// fullWidthReplacements maps full-width syntax characters to their half-width equivalents.
// Used for error messages, not auto-correction (fail-fast).
var fullWidthReplacements = map[rune]rune{
    '＃': '#',
    '＠': '@',
    '（': '(',
    '）': ')',
    '－': '-',
}

// validateNoFullWidthSymbols checks for full-width characters that should be half-width
// in Yahoo Realtime Search syntax.
func validateNoFullWidthSymbols(s string) error {
    for _, r := range s {
        if hw, ok := fullWidthReplacements[r]; ok {
            return ierrors.Wrapf(ErrInvalidParameter,
                "full-width character %q found, use half-width %q instead", string(r), string(hw))
        }
    }
    return nil
}

// validateNotEmpty checks that a string is not empty or whitespace-only.
func validateNotEmpty(field, value string) error {
    if strings.TrimSpace(value) == "" {
        return ierrors.Wrapf(ErrInvalidParameter, "%s cannot be empty", field)
    }
    return nil
}

// validateNoPrefix checks that a value does not start with its auto-added prefix.
// This prevents double-prefixing (e.g., "##tag", "ID:ID:user", "@@user").
func validateNoPrefix(field, value, prefix string) error {
    if strings.HasPrefix(value, prefix) {
        return ierrors.Wrapf(ErrInvalidParameter,
            "%s value should not include %q prefix: %q", field, prefix, value)
    }
    return nil
}
```

## Client API Addition

### New Method on Client

```go
// SearchWithQuery performs a search using a Query builder.
func (c *Client) SearchWithQuery(ctx context.Context, q *Query) (*SearchResult, error) {
    queryStr, err := q.Build()
    if err != nil {
        return nil, err
    }
    return c.Search(ctx, queryStr)
}

// SearchWithQueryAndLimit performs a search using a Query builder with a result limit.
func (c *Client) SearchWithQueryAndLimit(ctx context.Context, q *Query, limit int) (*SearchResult, error) {
    queryStr, err := q.Build()
    if err != nil {
        return nil, err
    }
    return c.SearchWithLimit(ctx, queryStr, limit)
}
```

The existing `Search(ctx, "raw query")` and `SearchWithLimit(ctx, "raw query", limit)` remain unchanged. Users who prefer raw query strings can continue using them.

## CLI Flag Additions

### New Flags for cmd/yrs/main.go

```go
&cli.StringSliceFlag{
    Name:  "not",
    Usage: "Exclude posts containing these keywords (NOT search)",
},
&cli.StringFlag{
    Name:  "from",
    Usage: "Filter posts from a specific account (ID:xxx)",
},
&cli.StringFlag{
    Name:  "to",
    Usage: "Filter posts mentioning a specific account (@xxx)",
},
&cli.StringSliceFlag{
    Name:  "hashtag",
    Usage: "Filter posts containing hashtags (#xxx)",
},
&cli.StringFlag{
    Name:  "url",
    Usage: "Filter posts containing a URL/domain (URL:xxx)",
},
&cli.StringSliceFlag{
    Name:  "or",
    Usage: "OR keywords group (posts containing any of these)",
},
```

### CLI Action Update

When any query builder flag is provided, construct a `Query` instead of using the raw positional argument directly:

```go
Action: func(ctx context.Context, cmd *cli.Command) error {
    query := cmd.Args().First()

    // Check if any query builder flags are set
    hasBuilderFlags := len(cmd.StringSlice("not")) > 0 ||
        cmd.String("from") != "" ||
        cmd.String("to") != "" ||
        len(cmd.StringSlice("hashtag")) > 0 ||
        cmd.String("url") != "" ||
        len(cmd.StringSlice("or")) > 0

    var queryStr string
    if hasBuilderFlags {
        qb := yrs.NewQuery()
        if query != "" {
            qb.And(query)
        }
        for _, kw := range cmd.StringSlice("not") {
            qb.Not(kw)
        }
        if from := cmd.String("from"); from != "" {
            qb.FromUser(from)
        }
        if to := cmd.String("to"); to != "" {
            qb.ToUser(to)
        }
        for _, tag := range cmd.StringSlice("hashtag") {
            qb.Hashtag(tag)
        }
        if u := cmd.String("url"); u != "" {
            qb.URL(u)
        }
        if orKeywords := cmd.StringSlice("or"); len(orKeywords) > 0 {
            qb.Or(orKeywords...)
        }
        var err error
        queryStr, err = qb.Build()
        if err != nil {
            return cli.Exit(fmt.Sprintf("Invalid query: %v", err), 1)
        }
    } else {
        queryStr = query
    }

    if queryStr == "" {
        return cli.Exit("Search query is required.", 1)
    }
    // ... rest of existing logic using queryStr
}
```

### CLI Usage Examples

```bash
# Raw query (unchanged)
yrs "golang tutorial"

# NOT search
yrs "golang" --not tutorial --not beginner

# From specific account
yrs --from YahooSearchJP

# Combined filters
yrs "golang" --not tutorial --from YahooSearchJP

# OR search
yrs --or 猫 --or 犬

# Hashtag filter
yrs --hashtag 地震

# URL filter
yrs "news" --url yahoo.co.jp

# Complex query
yrs "投資" --not "仮想通貨" --hashtag 株式投資 --from sbaboreki
```

## File Changes Summary

| File | Change |
|------|--------|
| `query.go` (new) | `Query` struct, `NewQuery()`, method chain, `Build()` |
| `validate.go` (new) | `validateNoFullWidthSymbols()`, `validateNotEmpty()` |
| `client.go` | Add `SearchWithQuery()`, `SearchWithQueryAndLimit()` |
| `cmd/yrs/main.go` | Add `--not`, `--from`, `--to`, `--hashtag`, `--url`, `--or` flags |
| `query_test.go` (new) | Unit tests for query building and validation |

## Test Strategy

### query_test.go

#### Build Output Tests

| Test Case | Input | Expected Output |
|-----------|-------|-----------------|
| Single keyword | `NewQuery("golang").Build()` | `"golang"` |
| Multiple AND | `NewQuery("golang", "tutorial").Build()` | `"golang tutorial"` |
| OR group | `NewQuery("猫").Or("犬", "うさぎ").Build()` | `"猫 (犬 うさぎ)"` |
| NOT | `NewQuery("golang").Not("tutorial").Build()` | `"golang -tutorial"` |
| FromUser | `NewQuery("news").FromUser("YahooSearchJP").Build()` | `"news ID:YahooSearchJP"` |
| ToUser | `NewQuery().ToUser("YahooSearchJP").Build()` | `"@YahooSearchJP"` |
| Hashtag | `NewQuery().Hashtag("地震").Build()` | `"#地震"` |
| URL filter | `NewQuery().URL("yahoo.co.jp").Build()` | `"URL:yahoo.co.jp"` |
| Complex | `NewQuery("golang").Not("tutorial").FromUser("user1").Hashtag("go").Build()` | `"golang -tutorial ID:user1 #go"` |
| Only OR | `NewQuery().Or("猫", "犬").Build()` | `"(猫 犬)"` |
| Multiple OR groups | `NewQuery().Or("猫", "犬").Or("赤", "青").Build()` | `"(猫 犬) (赤 青)"` |

#### Validation Error Tests

| Test Case | Input | Expected Error |
|-----------|-------|---------------|
| Empty query | `NewQuery().Build()` | `ErrInvalidParameter` (empty) |
| Full-width hash | `NewQuery().Hashtag("＃地震").Build()` | full-width error |
| Full-width at | `NewQuery().ToUser("＠user").Build()` | full-width error |
| Full-width minus in NOT | `NewQuery("test").Not("－bad").Build()` | full-width error |
| Empty FromUser | `NewQuery().FromUser("").Build()` | empty error |
| Hashtag with space | `NewQuery().Hashtag("hello world").Build()` | space in hashtag error |
| Hashtag with # prefix | `NewQuery().Hashtag("#地震").Build()` | leading prefix error |
| FromUser with ID: prefix | `NewQuery().FromUser("ID:user").Build()` | leading prefix error |
| ToUser with @ prefix | `NewQuery().ToUser("@user").Build()` | leading prefix error |

### Integration with Existing Tests

No changes to existing `client_test.go` or `scraper_test.go`. The query builder is purely a string-building layer tested independently.

## Package Structure (Updated)

```
go-yahoo-realtime-search/
├── client.go          -- Add SearchWithQuery, SearchWithQueryAndLimit
├── query.go           -- NEW: Query struct, NewQuery, Build
├── validate.go        -- NEW: full-width detection, input validation
├── query_test.go      -- NEW: unit tests for query builder
├── options.go         -- (unchanged)
├── config.go          -- (unchanged)
├── constants.go       -- (unchanged)
├── types.go           -- (unchanged)
├── scraper.go         -- (unchanged)
├── errors.go          -- (unchanged)
├── cmd/yrs/main.go    -- Add query builder flags
└── ...
```
