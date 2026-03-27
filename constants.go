package yrs

import "time"

const (
	// LibraryName is the name of this library.
	LibraryName = "go-yahoo-realtime-search"

	// LibraryVersion is the current version of this library.
	LibraryVersion = "0.1.0"

	// DefaultUserAgent is the default User-Agent header sent with HTTP requests.
	DefaultUserAgent = LibraryName + "/" + LibraryVersion

	// DefaultRequestTimeout is the default timeout for HTTP requests.
	DefaultRequestTimeout = 30 * time.Second

	// baseURL is the Yahoo Realtime Search endpoint.
	baseURL = "https://search.yahoo.co.jp/realtime/search"
)
