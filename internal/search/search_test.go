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

	var hits int
	for r := range results {
		if r.Err != nil {
			t.Fatal(r.Err)
		}
		hits++
	}
	if hits != 1 {
		t.Errorf("expected 1 match, got %d", hits)
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

	var hits int
	for r := range results {
		if r.Err != nil {
			t.Fatal(r.Err)
		}
		hits++
	}
	if hits != 0 {
		t.Errorf("expected 0 matches, got %d", hits)
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

	var hits int
	for r := range results {
		if r.Err != nil {
			t.Fatal(r.Err)
		}
		hits++
	}
	if hits != 2 {
		t.Errorf("expected 2 matches (one per root), got %d", hits)
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
		found = true
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
		hits++
	}
	if hits != 3 {
		t.Errorf("expected 3 regex matches, got %d", hits)
	}
}

func TestSearchNonexistentRoot(t *testing.T) {
	m := matcher.NewLiteral("hello", false)
	_, err := Search(context.Background(), []string{"/nonexistent/path"}, m, 2)
	if err == nil {
		t.Error("expected error for nonexistent root")
	}
}

func TestSearchSkipsBinary(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "binary.bin")
	if err := os.WriteFile(path, []byte{0x00, 0x01, 0x02}, 0644); err != nil {
		t.Fatal(err)
	}

	// Register a dummy extractor for .bin that opens raw (no binary detection).
	extract.Register(".bin", &rawExtractor{})

	m := matcher.NewLiteral("hello", false)
	results, err := Search(context.Background(), []string{dir}, m, 2)
	if err != nil {
		t.Fatal(err)
	}

	for r := range results {
		if r.Err != nil {
			t.Fatal(r.Err)
		}
		if r.Line != "" {
			t.Error("expected no matches in binary file")
		}
	}
}

type rawExtractor struct{}

func (t *rawExtractor) Extract(path string) (io.ReadCloser, error) {
	return os.Open(path)
}
