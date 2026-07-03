package extract

import "io"

// Extractor defines an interface for reading text from a file.
// The caller must close the returned reader.
// If the file is binary, Extract returns (nil, nil) — the caller should skip it.
// If the file cannot be read, Extract returns (nil, err).
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
