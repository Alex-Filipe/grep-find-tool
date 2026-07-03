package output

import (
	"fmt"
	"os"

	"github.com/user/grep-tool/internal/search"
)

// ColorMode controls when to use ANSI color codes.
type ColorMode int

const (
	ColorAuto ColorMode = iota
	ColorAlways
	ColorNever
)

// Formatter formats search results for display.
type Formatter struct {
	colorOut bool // use color for stdout (matches)
	colorErr bool // use color for stderr (errors)
}

// NewFormatter creates a Formatter with the given color mode.
// TTY detection is done once at construction time.
func NewFormatter(mode ColorMode) *Formatter {
	return &Formatter{
		colorOut: resolveColor(mode, os.Stdout),
		colorErr: resolveColor(mode, os.Stderr),
	}
}

// ColorModeFromFlag parses the --color flag value.
func ColorModeFromFlag(s string) ColorMode {
	switch s {
	case "always":
		return ColorAlways
	case "never":
		return ColorNever
	default:
		return ColorAuto
	}
}

// FormatResult renders a single search.Result as a string.
// Format: "path:line_num:content" with optional ANSI colors.
// Errors are formatted separately for stderr.
func (f *Formatter) FormatResult(r search.Result) string {
	if r.Err != nil {
		return f.formatErr(r)
	}
	return f.formatMatch(r)
}

func resolveColor(mode ColorMode, f *os.File) bool {
	switch mode {
	case ColorAlways:
		return true
	case ColorNever:
		return false
	default:
		fi, err := f.Stat()
		if err != nil {
			return false
		}
		return (fi.Mode() & os.ModeCharDevice) != 0
	}
}

func (f *Formatter) formatMatch(r search.Result) string {
	if f.colorOut {
		return fmt.Sprintf("\033[1;36m%s\033[0m:\033[1;33m%d\033[0m:%s", r.Path, r.LineNum, r.Line)
	}
	return fmt.Sprintf("%s:%d:%s", r.Path, r.LineNum, r.Line)
}

func (f *Formatter) formatErr(r search.Result) string {
	if f.colorErr {
		return fmt.Sprintf("\033[1;31m%s\033[0m: %v", r.Path, r.Err)
	}
	return fmt.Sprintf("%s: %v", r.Path, r.Err)
}
