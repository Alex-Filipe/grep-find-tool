package cmd

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/user/grep-tool/internal/walker"
)

func init() {
	rootCmd.AddCommand(findCmd)
}

var findCmd = &cobra.Command{
	Use:   "find <name> [paths...]",
	Short: "Search for files by name",
	Long: `Search for files matching the given name pattern.
Supports glob-style patterns like "*.go".

Examples:
  grep-tool find "*.go"
  grep-tool find "test_*" src/`,
	Args: cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		pattern := args[0]
		paths := args[1:]
		if len(paths) == 0 {
			paths = []string{"."}
		}

		if _, err := filepath.Match(pattern, "x"); err != nil {
			return fmt.Errorf("invalid pattern %q: %w", pattern, err)
		}

		ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
		defer stop()

		for _, root := range paths {
			ch, err := walker.Walk(ctx, root)
			if err != nil {
				fmt.Fprintf(os.Stderr, "%s: %v\n", root, err)
				continue
			}
			for p := range ch {
				if matched, _ := filepath.Match(pattern, filepath.Base(p)); matched {
					fmt.Println(p)
				}
			}
		}

		return nil
	},
}
