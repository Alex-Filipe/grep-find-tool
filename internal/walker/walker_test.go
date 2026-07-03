package walker

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"testing"
)

func TestWalkSingleFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "file.txt")
	if err := os.WriteFile(path, []byte("hello"), 0644); err != nil {
		t.Fatal(err)
	}

	ch, err := Walk(context.Background(), path)
	if err != nil {
		t.Fatal(err)
	}
	var paths []string
	for p := range ch {
		paths = append(paths, p)
	}
	if len(paths) != 1 || paths[0] != path {
		t.Errorf("expected [%s], got %v", path, paths)
	}
}

func TestWalkDirectory(t *testing.T) {
	dir := t.TempDir()
	files := []string{"a.txt", "b.go", "c.md"}
	for _, f := range files {
		if err := os.WriteFile(filepath.Join(dir, f), []byte(f), 0644); err != nil {
			t.Fatal(err)
		}
	}

	ch, err := Walk(context.Background(), dir)
	if err != nil {
		t.Fatal(err)
	}
	var paths []string
	for p := range ch {
		paths = append(paths, p)
	}
	sort.Strings(paths)
	if len(paths) != len(files) {
		t.Errorf("expected %d files, got %d: %v", len(files), len(paths), paths)
	}
}

func TestWalkSkipsHidden(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "visible.txt"), []byte("ok"), 0644); err != nil {
		t.Fatal(err)
	}
	hiddenDir := filepath.Join(dir, ".hidden")
	if err := os.Mkdir(hiddenDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(hiddenDir, "secret.txt"), []byte("secret"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, ".dotfile"), []byte("dot"), 0644); err != nil {
		t.Fatal(err)
	}

	ch, err := Walk(context.Background(), dir)
	if err != nil {
		t.Fatal(err)
	}
	var paths []string
	for p := range ch {
		paths = append(paths, p)
	}
	if len(paths) != 1 {
		t.Errorf("expected 1 visible file, got %d: %v", len(paths), paths)
	}
}

func TestWalkSkipsGit(t *testing.T) {
	dir := t.TempDir()
	gitDir := filepath.Join(dir, ".git")
	if err := os.Mkdir(gitDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(gitDir, "config"), []byte("repo"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "main.go"), []byte("package main"), 0644); err != nil {
		t.Fatal(err)
	}

	ch, err := Walk(context.Background(), dir)
	if err != nil {
		t.Fatal(err)
	}
	var paths []string
	for p := range ch {
		paths = append(paths, p)
	}
	if len(paths) != 1 {
		t.Errorf("expected 1 file (not counting .git), got %d: %v", len(paths), paths)
	}
}

func TestWalkCancellation(t *testing.T) {
	dir := t.TempDir()
	for i := 0; i < 100; i++ {
		name := filepath.Join(dir, fmt.Sprintf("file%d.txt", i))
		if err := os.WriteFile(name, []byte("data"), 0644); err != nil {
			t.Fatal(err)
		}
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	ch, err := Walk(ctx, dir)
	if err != nil {
		t.Fatal(err)
	}
	count := 0
	for range ch {
		count++
	}
	if count > 0 {
		t.Errorf("expected 0 files with cancelled context, got %d", count)
	}
}

func TestWalkHiddenRoot(t *testing.T) {
	dir := t.TempDir()
	hiddenDir := filepath.Join(dir, ".config")
	if err := os.Mkdir(hiddenDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(hiddenDir, "settings.json"), []byte("{}"), 0644); err != nil {
		t.Fatal(err)
	}

	ch, err := Walk(context.Background(), hiddenDir)
	if err != nil {
		t.Fatal(err)
	}
	var paths []string
	for p := range ch {
		paths = append(paths, p)
	}
	if len(paths) != 1 {
		t.Errorf("expected 1 file inside hidden root, got %d: %v", len(paths), paths)
	}
}

func TestWalkNonexistentRoot(t *testing.T) {
	_, err := Walk(context.Background(), "/nonexistent/path")
	if err == nil {
		t.Error("expected error for nonexistent root")
	}
}
