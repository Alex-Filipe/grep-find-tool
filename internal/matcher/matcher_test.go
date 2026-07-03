package matcher

import (
	"testing"
)

func TestLiteralCaseSensitive(t *testing.T) {
	m := NewLiteral("hello", false)
	tests := []struct {
		line    string
		want    bool
	}{
		{"hello world", true},
		{"Hello world", false},
		{"say hello", true},
		{"hella", false},
		{"", false},
		{"HELLO", false},
		{"hello", true},
	}
	for _, tt := range tests {
		if got := m(tt.line); got != tt.want {
			t.Errorf("NewLiteral(%q, false)(%q) = %v, want %v", "hello", tt.line, got, tt.want)
		}
	}
}

func TestLiteralCaseInsensitive(t *testing.T) {
	m := NewLiteral("hello", true)
	tests := []struct {
		line    string
		want    bool
	}{
		{"hello world", true},
		{"Hello world", true},
		{"HELLO", true},
		{"hElLo", true},
		{"hella", false},
		{"", false},
		{"world", false},
	}
	for _, tt := range tests {
		if got := m(tt.line); got != tt.want {
			t.Errorf("NewLiteral(%q, true)(%q) = %v, want %v", "hello", tt.line, got, tt.want)
		}
	}
}

func TestLiteralEmptyPattern(t *testing.T) {
	m := NewLiteral("", false)
	if !m("anything") {
		t.Error("empty literal pattern should match everything")
	}
	if !m("") {
		t.Error("empty literal pattern should match empty string")
	}
}

func TestLiteralUTF8(t *testing.T) {
	m := NewLiteral("coração", false)
	if !m("meu coração bate forte") {
		t.Error("should match UTF-8 substring")
	}
	if m("meu coraCAO bate forte") {
		t.Error("should not match different casing")
	}

	mFold := NewLiteral("coração", true)
	if !mFold("meu CORAÇÃO bate forte") {
		t.Error("case-insensitive should match UTF-8 uppercase")
	}
}

func TestRegexCaseSensitive(t *testing.T) {
	m, err := NewRegex(`h.llo`, false)
	if err != nil {
		t.Fatal(err)
	}
	tests := []struct {
		line    string
		want    bool
	}{
		{"hello", true},
		{"hxllo", true},
		{"h llo", true},
		{"Hello", false},
		{"hllo", false},
		{"", false},
	}
	for _, tt := range tests {
		if got := m(tt.line); got != tt.want {
			t.Errorf("NewRegex(%q, false)(%q) = %v, want %v", `h.llo`, tt.line, got, tt.want)
		}
	}
}

func TestRegexCaseInsensitive(t *testing.T) {
	m, err := NewRegex(`h.llo`, true)
	if err != nil {
		t.Fatal(err)
	}
	if !m("Hello") {
		t.Error("case-insensitive regex should match 'Hello'")
	}
	if !m("hxllo") {
		t.Error("case-insensitive regex should match 'hxllo'")
	}
}

func TestRegexInvalid(t *testing.T) {
	_, err := NewRegex(`[invalid`, false)
	if err == nil {
		t.Fatal("expected error for invalid regex")
	}
}

func TestRegexAnchored(t *testing.T) {
	m, err := NewRegex(`^hello`, false)
	if err != nil {
		t.Fatal(err)
	}
	if !m("hello world") {
		t.Error("anchored regex should match at start")
	}
	if m("say hello") {
		t.Error("anchored regex should not match in middle")
	}
}

func TestRegexEmptyPattern(t *testing.T) {
	m, err := NewRegex(``, false)
	if err != nil {
		t.Fatal(err)
	}
	if !m("anything") {
		t.Error("empty regex should match everything")
	}
}
