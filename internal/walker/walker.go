package walker

import "context"

// Walk traverses a directory tree and sends file paths
// into the returned channel. It skips .git/ and hidden directories.
// The context allows cancellation (e.g. Ctrl-C).
// TODO: honor .gitignore patterns.
func Walk(ctx context.Context, root string) (<-chan string, error) {
	return nil, nil
}
