package extract

import (
	"os"
	"path/filepath"
	"testing"
)

func TestTextExtractorFound(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.go")
	if err := os.WriteFile(path, []byte("package main\n"), 0644); err != nil {
		t.Fatal(err)
	}

	e := For(".go")
	if e == nil {
		t.Fatal("expected extractor for .go")
	}
	r, err := e.Extract(path)
	if err != nil {
		t.Fatal(err)
	}
	if r == nil {
		t.Fatal("expected non-nil reader")
	}
	r.Close()
}

func TestTextExtractorNotFound(t *testing.T) {
	e := For(".xyz")
	if e != nil {
		t.Error("expected no extractor for .xyz")
	}
}

func TestBinaryDetection(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "binary.bin")
	data := []byte{0x00, 0x01, 0x02, 0x03}
	if err := os.WriteFile(path, data, 0644); err != nil {
		t.Fatal(err)
	}

	if !isBinary(path) {
		t.Error("expected binary detection for file with NUL byte")
	}
}

func TestNonBinaryDetection(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "text.txt")
	data := []byte("hello world\nthis is text\n")
	if err := os.WriteFile(path, data, 0644); err != nil {
		t.Fatal(err)
	}

	if isBinary(path) {
		t.Error("expected non-binary for plain text")
	}
}

func TestBinaryFileSkippedByExtractor(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "binary.go")
	data := []byte("package main\x00\n")
	if err := os.WriteFile(path, data, 0644); err != nil {
		t.Fatal(err)
	}

	e := For(".go")
	if e == nil {
		t.Fatal("expected extractor for .go")
	}
	r, err := e.Extract(path)
	if err != nil {
		t.Fatal(err)
	}
	if r != nil {
		t.Error("expected nil reader for binary file")
	}
}

func TestPDFExtractor(t *testing.T) {
	e := For(".pdf")
	if e == nil {
		t.Fatal("expected extractor for .pdf")
	}
	r, err := e.Extract("/nonexistent/file.pdf")
	if err != nil {
		t.Fatal(err)
	}
	if r != nil {
		t.Error("expected nil reader for .pdf stub")
	}
}

func TestEmptyFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "empty.go")
	if err := os.WriteFile(path, []byte{}, 0644); err != nil {
		t.Fatal(err)
	}

	e := For(".go")
	r, err := e.Extract(path)
	if err != nil {
		t.Fatal(err)
	}
	if r == nil {
		t.Fatal("expected non-nil reader for empty file")
	}
	r.Close()
}
