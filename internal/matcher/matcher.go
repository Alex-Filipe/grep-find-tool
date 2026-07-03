package matcher

// MatchFunc is a function that checks if a line matches the pattern.
type MatchFunc func(line string) bool

// NewLiteral creates a MatchFunc for literal substring matching.
func NewLiteral(pattern string, ignoreCase bool) MatchFunc {
	return nil
}

// NewRegex creates a MatchFunc for regex matching.
func NewRegex(pattern string, ignoreCase bool) (MatchFunc, error) {
	return nil, nil
}
