package matcher

import (
	"regexp"
	"strings"
)

// MatchFunc is a function that checks if a line matches the pattern.
type MatchFunc func(line string) bool

// NewLiteral creates a MatchFunc for literal substring matching.
// When ignoreCase is true, matching is case-insensitive.
func NewLiteral(pattern string, ignoreCase bool) MatchFunc {
	if ignoreCase {
		lowerPattern := strings.ToLower(pattern)
		return func(line string) bool {
			// alloc: ToLower cria string nova por linha.
			// Se for gargalo, migrar para bytes-based ou scan rune a rune.
			return strings.Contains(strings.ToLower(line), lowerPattern)
		}
	}
	return func(line string) bool {
		return strings.Contains(line, pattern)
	}
}

// NewRegex creates a MatchFunc for regex pattern matching.
// When ignoreCase is true, the regex is compiled with the (?i) flag.
func NewRegex(pattern string, ignoreCase bool) (MatchFunc, error) {
	if ignoreCase {
		pattern = "(?i)" + pattern
	}
	re, err := regexp.Compile(pattern)
	if err != nil {
		return nil, err
	}
	return func(line string) bool {
		return re.MatchString(line)
	}, nil
}
