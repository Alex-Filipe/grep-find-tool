package search

import (
	"context"

	"github.com/user/grep-tool/internal/matcher"
)

// Result holds a single match or error found during search.
// If Err != nil, Path names the file that failed and LineNum/Line are zero.
type Result struct {
	Path    string
	LineNum int
	Line    string
	Err     error
}

// Search runs the full pipeline: walker -> extract -> matcher.
// Uses a worker pool for parallel processing over all given roots.
// Results are streamed incrementally through the returned channel (unordered).
// The returned error is only for setup failures (e.g. root doesn't exist);
// per-file errors are embedded in Result.Err.
func Search(ctx context.Context, roots []string, match matcher.MatchFunc, workers int) (<-chan Result, error) {
	return nil, nil
}
