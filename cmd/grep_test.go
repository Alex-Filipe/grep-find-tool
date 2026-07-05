package cmd

import (
	"os"
	"path/filepath"
	"sort"
	"testing"
)

func TestBuildMatcher(t *testing.T) {
	t.Run("literal treats metachars literally", func(t *testing.T) {
		m, err := buildMatcher("a.b", false, false)
		if err != nil {
			t.Fatal(err)
		}
		if !m("a.b") {
			t.Error(`literal "a.b" should match "a.b"`)
		}
		if m("axb") {
			t.Error(`literal "a.b" should NOT match "axb"`)
		}
	})

	t.Run("regex mode interprets metachars", func(t *testing.T) {
		m, err := buildMatcher("a.b", false, true)
		if err != nil {
			t.Fatal(err)
		}
		if !m("axb") {
			t.Error(`regex "a.b" should match "axb"`)
		}
	})

	t.Run("ignoreCase applies", func(t *testing.T) {
		m, err := buildMatcher("hello", true, false)
		if err != nil {
			t.Fatal(err)
		}
		if !m("HELLO") {
			t.Error("case-insensitive literal should match HELLO")
		}
	})

	t.Run("invalid regex returns error", func(t *testing.T) {
		if _, err := buildMatcher("[abc", false, true); err == nil {
			t.Error("expected error for invalid regex")
		}
	})
}

func TestOrDot(t *testing.T) {
	if got := orDot(""); got != "." {
		t.Errorf(`orDot("") = %q, want "."`, got)
	}
	if got := orDot("src"); got != "src" {
		t.Errorf(`orDot("src") = %q, want "src"`, got)
	}
}

func TestSubdirsOf(t *testing.T) {
	dir := t.TempDir()
	for _, sub := range []string{"alpha", "beta", ".hidden"} {
		if err := os.Mkdir(filepath.Join(dir, sub), 0755); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.WriteFile(filepath.Join(dir, "file.txt"), []byte("x"), 0644); err != nil {
		t.Fatal(err)
	}

	got, err := subdirsOf(dir)
	if err != nil {
		t.Fatal(err)
	}
	sort.Strings(got)

	want := []string{"alpha", "beta"} // hidden dir and plain file excluded
	if len(got) != len(want) || got[0] != want[0] || got[1] != want[1] {
		t.Errorf("subdirsOf = %v, want %v", got, want)
	}
}
