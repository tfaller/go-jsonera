// Package jsonp provides functions to parse and format RFC RFC6901 JSON Pointer
package jsonp

import (
	"errors"
	"strings"
)

// ErrMissingTokenPrefix indicates that a "/" was missing
var ErrMissingTokenPrefix = errors.New("Missing token prefix")

var decoder = strings.NewReplacer("~1", "/", "~0", "~")
var encoder = strings.NewReplacer("~", "~0", "/", "~1")

// Parse parses a RFC6901 JSON Pointer into its parts
func Parse(pointer string) ([]string, error) {
	if pointer == "" {
		return []string{}, nil
	}
	if pointer[0] != '/' {
		return nil, ErrMissingTokenPrefix
	}
	parts := strings.Split(pointer[1:], "/")
	for i, p := range parts {
		parts[i] = decoder.Replace(p)
	}
	return parts, nil
}

// Format formats a pointer to a RFC6901 JSON Pointer
func Format(pointer []string) string {
	s := strings.Builder{}
	for _, part := range pointer {
		s.WriteRune('/')
		s.WriteString(encoder.Replace(part))
	}
	return s.String()
}
