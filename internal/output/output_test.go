package output

import (
	"errors"
	"testing"

	"github.com/user/grep-tool/internal/search"
)

func TestFormatMatchNoColor(t *testing.T) {
	f := NewFormatter(ColorNever)
	r := search.Result{Path: "a.txt", LineNum: 3, Line: "hello world"}
	got := f.FormatResult(r)
	want := "a.txt:3:hello world"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestFormatErrNoColor(t *testing.T) {
	f := NewFormatter(ColorNever)
	r := search.Result{Path: "a.txt", Err: errors.New("permission denied")}
	got := f.FormatResult(r)
	want := "a.txt: permission denied"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestFormatMatchAlwaysColor(t *testing.T) {
	f := NewFormatter(ColorAlways)
	r := search.Result{Path: "a.txt", LineNum: 3, Line: "hello world"}
	got := f.FormatResult(r)
	if len(got) < 10 {
		t.Errorf("expected ANSI escapes, got short string %q", got)
	}
	if got[0] != '\033' {
		t.Errorf("expected ANSI escape (\\033), got %q", got[:1])
	}
}

func TestFormatErrAlwaysColor(t *testing.T) {
	f := NewFormatter(ColorAlways)
	r := search.Result{Path: "a.txt", Err: errors.New("permission denied")}
	got := f.FormatResult(r)
	if got[0] != '\033' {
		t.Errorf("expected ANSI escape for error, got %q", got[:1])
	}
}

func TestFormatEmptyLine(t *testing.T) {
	f := NewFormatter(ColorNever)
	r := search.Result{Path: "a.txt", LineNum: 0, Line: ""}
	got := f.FormatResult(r)
	if got != "" {
		t.Errorf("expected empty string for empty line, got %q", got)
	}
}

func TestColorModeFromFlag(t *testing.T) {
	tests := []struct {
		flag string
		want ColorMode
	}{
		{"auto", ColorAuto},
		{"always", ColorAlways},
		{"never", ColorNever},
		{"invalid", ColorAuto},
	}
	for _, tt := range tests {
		got := ColorModeFromFlag(tt.flag)
		if got != tt.want {
			t.Errorf("ColorModeFromFlag(%q) = %d, want %d", tt.flag, got, tt.want)
		}
	}
}
