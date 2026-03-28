package yrs

import (
	"strings"

	ierrors "github.com/cnosuke/go-yahoo-realtime-search/internal/errors"
)

// fullWidthReplacements maps full-width syntax characters to their half-width equivalents.
var fullWidthReplacements = map[rune]rune{
	'＃': '#',
	'＠': '@',
	'（': '(',
	'）': ')',
	'－': '-',
}

func validateNoFullWidthSymbols(s string) error {
	for _, r := range s {
		if hw, ok := fullWidthReplacements[r]; ok {
			return ierrors.Wrapf(ErrInvalidParameter,
				"full-width character %q found, use half-width %q instead", string(r), string(hw))
		}
	}
	return nil
}

func validateNotEmpty(field, value string) error {
	if strings.TrimSpace(value) == "" {
		return ierrors.Wrapf(ErrInvalidParameter, "%s cannot be empty", field)
	}
	return nil
}

func validateNoPrefix(field, value, prefix string) error {
	if strings.HasPrefix(value, prefix) {
		return ierrors.Wrapf(ErrInvalidParameter,
			"%s value should not include %q prefix: %q", field, prefix, value)
	}
	return nil
}
