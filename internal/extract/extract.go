package extract

import (
	"io"
	"os"
)

// Extractor defines an interface for reading text from a file.
// The caller must close the returned reader.
// If the file cannot be read or is binary, Extract returns (nil, nil)
// and the caller should skip it.
type Extractor interface {
	Extract(path string) (io.ReadCloser, error)
}

// registry maps file extensions to extractors.
// Only written during init() calls — safe for concurrent reads.
var registry = make(map[string]Extractor)

// Register binds an extractor to one or more extensions.
// Intended for use in init(); not safe for concurrent runtime calls.
func Register(ext string, e Extractor) {
	registry[ext] = e
}

// For returns the extractor registered for the given extension.
func For(ext string) Extractor {
	return registry[ext]
}

// isBinary checks if a file appears to be binary by looking for NUL bytes
// in the first 512 bytes.
func isBinary(path string) bool {
	f, err := os.Open(path)
	if err != nil {
		return false
	}
	defer f.Close()

	buf := make([]byte, 512)
	n, err := f.Read(buf)
	if err != nil && n == 0 {
		return false
	}
	for _, b := range buf[:n] {
		if b == 0 {
			return true
		}
	}
	return false
}
