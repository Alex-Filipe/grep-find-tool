package cmd

import (
	"errors"
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

// ErrNoMatches is returned by grep when no matches are found.
// It signals exit code 1 (POSIX convention: 1 = no matches).
var ErrNoMatches = errors.New("no matches found")

var (
	workers int
)

var rootCmd = &cobra.Command{
	Use:   "grep-tool",
	Short: "A fast text search tool for files",
	Long: `grep-tool searches for text inside files or finds files by name.

Usage:
  grep-tool grep <pattern> [paths...]
  grep-tool find <name> [paths...]`,
	SilenceUsage:  true,
	SilenceErrors: true,
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		if errors.Is(err, ErrNoMatches) {
			os.Exit(1)
		}
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(2)
	}
}

func init() {
	rootCmd.PersistentFlags().IntVarP(&workers, "workers", "j", 4, "number of parallel workers")
}
