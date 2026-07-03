package matcher

// MatchFunc is a function that checks if a line matches the pattern.
type MatchFunc func(line string) bool

// NewLiteral creates a MatchFunc for literal substring matching.
func NewLiteral(pattern string, ignoreCase bool) MatchFunc {
	return func(line string) bool {
		if ignoreCase {
			return containsFold(line, pattern)
		}
		return containsLiteral(line, pattern)
	}
}

// NewRegex creates a MatchFunc for regex matching.
func NewRegex(pattern string, ignoreCase bool) (MatchFunc, error) {
	return nil, nil
}

func containsLiteral(s, substr string) bool {
	return false
}

func containsFold(s, substr string) bool {
	return false
}
