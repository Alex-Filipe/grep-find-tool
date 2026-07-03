package walker

import (
	"context"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

// Walk traverses a directory tree and sends file paths
// into the returned channel. It skips .git/ and hidden directories.
// The context allows cancellation (e.g. Ctrl-C).
// TODO: honor .gitignore patterns.
func Walk(ctx context.Context, root string) (<-chan string, error) {
	fi, err := os.Stat(root)
	if err != nil {
		return nil, err
	}
	if !fi.IsDir() {
		// Single file: send it directly.
		ch := make(chan string, 1)
		ch <- root
		close(ch)
		return ch, nil
	}

	ch := make(chan string)
	go func() {
		defer close(ch)
		filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				// Permission error or similar — skip and continue.
				return nil
			}

			select {
			case <-ctx.Done():
				return ctx.Err()
			default:
			}

			if d.IsDir() {
				name := d.Name()
				// Skip .git and hidden directories.
				if name == ".git" || (strings.HasPrefix(name, ".") && name != ".") {
					return filepath.SkipDir
				}
				return nil
			}

			// Skip hidden files.
			if strings.HasPrefix(d.Name(), ".") {
				return nil
			}

			select {
			case ch <- path:
			case <-ctx.Done():
				return ctx.Err()
			}

			return nil
		})
	}()
	return ch, nil
}
