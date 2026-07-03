package cmd

import (
	"context"
	"fmt"
	"os"
	"os/signal"

	"github.com/spf13/cobra"
	"github.com/user/grep-tool/internal/matcher"
	"github.com/user/grep-tool/internal/output"
	"github.com/user/grep-tool/internal/search"
)

var (
	ignoreCase bool
	regexpMode bool
	colorMode  string
)

func init() {
	rootCmd.AddCommand(grepCmd)
	grepCmd.Flags().BoolVarP(&ignoreCase, "ignore-case", "i", false, "case-insensitive search")
	grepCmd.Flags().BoolVarP(&regexpMode, "regexp", "e", false, "pattern is a regular expression")
	grepCmd.Flags().StringVar(&colorMode, "color", "auto", "use colors: auto, always, never")
}

var grepCmd = &cobra.Command{
	Use:   "grep <pattern> [paths...]",
	Short: "Search for text inside files",
	Long: `Search recursively for files containing the given pattern.

Pattern is treated as a literal substring by default.
Use --regexp (-e) to interpret the pattern as a regular expression.

Examples:
  grep-tool grep "TODO" src/
  grep-tool grep -i "error" .
  grep-tool grep -e "h.llo" .`,
	Args: cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		pattern := args[0]
		paths := args[1:]
		if len(paths) == 0 {
			paths = []string{"."}
		}

		var match matcher.MatchFunc
		if regexpMode {
			var err error
			match, err = matcher.NewRegex(pattern, ignoreCase)
			if err != nil {
				return fmt.Errorf("invalid regex: %w", err)
			}
		} else {
			match = matcher.NewLiteral(pattern, ignoreCase)
		}

		ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
		defer stop()

		formatter := output.NewFormatter(output.ColorModeFromFlag(colorMode))
		results, err := search.Search(ctx, paths, match, workers)
		if err != nil {
			return err
		}

		for r := range results {
			line := formatter.FormatResult(r)
			if r.Err != nil {
				fmt.Fprintln(os.Stderr, line)
			} else if line != "" {
				fmt.Println(line)
			}
		}

		return nil
	},
}
