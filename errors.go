package yrs

import "errors"

var (
	// ErrScrapeFailed is returned when the HTML scraping fails
	// (e.g., __NEXT_DATA__ not found, unexpected structure, non-200 HTTP status).
	ErrScrapeFailed = errors.New("yrs: scraping failed")

	// ErrInvalidParameter is returned for invalid input parameters.
	ErrInvalidParameter = errors.New("yrs: invalid parameter")
)
