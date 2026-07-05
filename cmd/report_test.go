package cmd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestSaveReport(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	content := "linha 1\nlinha 2\n"
	path, err := saveReport("busca", content)
	if err != nil {
		t.Fatal(err)
	}

	if filepath.Dir(path) != home {
		t.Errorf("report saved in %q, want under %q", path, home)
	}
	if !strings.HasSuffix(path, ".txt") {
		t.Errorf("report should be a .txt file, got %q", path)
	}

	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != content {
		t.Errorf("report content = %q, want %q", got, content)
	}
}

func TestBuildFindReport(t *testing.T) {
	got := buildFindReport("*.go", "/proj", []string{"/proj/a.go", "/proj/b.go"})

	for _, want := range []string{"Padrão:   *.go", "Pasta:    /proj", "Total:    2 arquivo(s)", "/proj/a.go", "/proj/b.go"} {
		if !strings.Contains(got, want) {
			t.Errorf("report missing %q\n---\n%s", want, got)
		}
	}
}

func TestReportsDirPrefersDownloads(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	downloads := filepath.Join(home, "Downloads")
	if err := os.Mkdir(downloads, 0755); err != nil {
		t.Fatal(err)
	}

	if got := reportsDir(); got != downloads {
		t.Errorf("reportsDir = %q, want %q", got, downloads)
	}
}
