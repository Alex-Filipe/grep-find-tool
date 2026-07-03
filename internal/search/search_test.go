package search

import (
	"context"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/user/grep-tool/internal/extract"
	"github.com/user/grep-tool/internal/matcher"
)

func TestSearchFound(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "a.txt"), []byte("hello world\nfoo bar\n"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "b.txt"), []byte("foo bar\nbaz qux\n"), 0644); err != nil {
		t.Fatal(err)
	}

	m := matcher.NewLiteral("hello", false)
	results, err := Search(context.Background(), []string{dir}, m, 2)
	if err != nil {
		t.Fatal(err)
	}

	var hits []Result
	for r := range results {
		if r.Err != nil {
			t.Fatal(r.Err)
		}
		if r.Line != "" {
			hits = append(hits, r)
		}
	}
	if len(hits) != 1 {
		t.Errorf("expected 1 match, got %d", len(hits))
	}
}

func TestSearchNoMatches(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "a.txt"), []byte("hello world\n"), 0644); err != nil {
		t.Fatal(err)
	}

	m := matcher.NewLiteral("nonexistent", false)
	results, err := Search(context.Background(), []string{dir}, m, 2)
	if err != nil {
		t.Fatal(err)
	}

	var hits []Result
	for r := range results {
		if r.Err != nil {
			t.Fatal(r.Err)
		}
		if r.Line != "" {
			hits = append(hits, r)
		}
	}
	if len(hits) != 0 {
		t.Errorf("expected 0 matches, got %d", len(hits))
	}
}

func TestSearchMultipleRoots(t *testing.T) {
	dirA := t.TempDir()
	dirB := t.TempDir()

	if err := os.WriteFile(filepath.Join(dirA, "a.txt"), []byte("target\n"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dirB, "b.txt"), []byte("target\n"), 0644); err != nil {
		t.Fatal(err)
	}

	m := matcher.NewLiteral("target", false)
	results, err := Search(context.Background(), []string{dirA, dirB}, m, 2)
	if err != nil {
		t.Fatal(err)
	}

	var hits []Result
	for r := range results {
		if r.Err != nil {
			t.Fatal(r.Err)
		}
		if r.Line != "" {
			hits = append(hits, r)
		}
	}
	if len(hits) != 2 {
		t.Errorf("expected 2 matches (one per root), got %d", len(hits))
	}
}

func TestSearchCaseInsensitive(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "a.txt"), []byte("Hello World\n"), 0644); err != nil {
		t.Fatal(err)
	}

	m := matcher.NewLiteral("hello", true)
	results, err := Search(context.Background(), []string{dir}, m, 2)
	if err != nil {
		t.Fatal(err)
	}

	found := false
	for r := range results {
		if r.Err != nil {
			t.Fatal(r.Err)
		}
		if r.Line != "" {
			found = true
		}
	}
	if !found {
		t.Error("expected case-insensitive match")
	}
}

func TestSearchRegex(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "a.txt"), []byte("hello\nhxllo\nhallo\n"), 0644); err != nil {
		t.Fatal(err)
	}

	m, err := matcher.NewRegex(`h.llo`, false)
	if err != nil {
		t.Fatal(err)
	}
	results, err := Search(context.Background(), []string{dir}, m, 2)
	if err != nil {
		t.Fatal(err)
	}

	var hits int
	for r := range results {
		if r.Err != nil {
			t.Fatal(r.Err)
		}
		if r.Line != "" {
			hits++
		}
	}
	if hits != 3 {
		t.Errorf("expected 3 regex matches, got %d", hits)
	}
}

func TestSearchSkipsBinary(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "binary.bin")
	if err := os.WriteFile(path, []byte{0x00, 0x01, 0x02}, 0644); err != nil {
		t.Fatal(err)
	}

	// Register a dummy extractor for .bin
	extract.Register(".bin", &textExtractor{})

	m := matcher.NewLiteral("hello", false)
	results, err := Search(context.Background(), []string{dir}, m, 2)
	if err != nil {
		t.Fatal(err)
	}

	hasLine := false
	for r := range results {
		if r.Err != nil {
			t.Fatal(r.Err)
		}
		if r.Line != "" {
			hasLine = true
		}
	}
	if hasLine {
		t.Error("expected no matches in binary file")
	}
}

type textExtractor struct{}

func (t *textExtractor) Extract(path string) (io.ReadCloser, error) {
	return os.Open(path)
}
