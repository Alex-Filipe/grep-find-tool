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

// ANSI style codes — the shared CLI palette. Kept in one place so the
// grep/find output and the interactive UI stay visually consistent.
const (
	Reset  = "\033[0m"
	Bold   = "\033[1m"
	Accent = "\033[38;5;208m" // orange — brand accent (headers, separators)
	Cyan   = "\033[1;36m"     // paths, info
	Yellow = "\033[1;33m"     // line numbers, warnings
	Red    = "\033[1;31m"     // errors
)

// Separator is the horizontal rule used across the tool's output.
const Separator = "─────────────────────────────────────"

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
		return fmt.Sprintf("%s%s%s:%s%d%s:%s", Cyan, r.Path, Reset, Yellow, r.LineNum, Reset, r.Line)
	}
	return fmt.Sprintf("%s:%d:%s", r.Path, r.LineNum, r.Line)
}

func (f *Formatter) formatErr(r search.Result) string {
	if f.colorErr {
		return fmt.Sprintf("%s%s%s: %v", Red, r.Path, Reset, r.Err)
	}
	return fmt.Sprintf("%s: %v", r.Path, r.Err)
}

// FormatGrouped renders results grouped by file with visual blocks.
// Each block has a file header with separator lines and indented matches.
// Requires results to be sorted by path (done by the caller).
func (f *Formatter) FormatGrouped(results []search.Result) string {
	if len(results) == 0 {
		return ""
	}

	var (
		buf       []byte
		sep       string
		headerFmt string
		lineFmt   string
	)

	if f.colorOut {
		sep = Accent + Separator + Reset + "\n"
		headerFmt = Accent + "%s" + Reset + "\n"
		lineFmt = "  " + Yellow + "Linha %d" + Reset + "    %s\n"
	} else {
		sep = Separator + "\n"
		headerFmt = "%s\n"
		lineFmt = "  Linha %d    %s\n"
	}

	currentPath := ""
	for _, r := range results {
		if r.Err != nil {
			continue
		}
		if r.Path != currentPath {
			if currentPath != "" {
				buf = append(buf, '\n')
			}
			buf = append(buf, []byte(sep)...)
			buf = append(buf, []byte(fmt.Sprintf(headerFmt, r.Path))...)
			buf = append(buf, []byte(sep)...)
			currentPath = r.Path
		}
		buf = append(buf, []byte(fmt.Sprintf(lineFmt, r.LineNum, r.Line))...)
	}

	return string(buf)
}
