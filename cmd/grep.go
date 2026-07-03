package cmd

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"sort"

	"github.com/spf13/cobra"
	"github.com/user/grep-tool/internal/matcher"
	"github.com/user/grep-tool/internal/output"
	"github.com/user/grep-tool/internal/search"
)

var (
	ignoreCase bool
	regexpMode bool
	colorMode  string
	sortMode   bool
)

func init() {
	rootCmd.AddCommand(grepCmd)
	grepCmd.Flags().BoolVarP(&ignoreCase, "ignore-case", "i", false, "case-insensitive search")
	grepCmd.Flags().BoolVarP(&regexpMode, "regexp", "e", false, "pattern is a regular expression")
	grepCmd.Flags().StringVar(&colorMode, "color", "auto", "use colors: auto, always, never")
	grepCmd.Flags().BoolVar(&sortMode, "sort", false, "group results by file (uses more memory)")
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

		if sortMode {
			matched, err := sortAndPrint(ctx, results, formatter)
			if err != nil {
				return err
			}
			if matched == 0 {
				return ErrNoMatches
			}
			return nil
		}

		matched := 0
		for r := range results {
			line := formatter.FormatResult(r)
			if r.Err != nil {
				fmt.Fprintln(os.Stderr, line)
			} else {
				fmt.Println(line)
				matched++
			}
		}

		if matched == 0 {
			return ErrNoMatches
		}
		return nil
	},
}

func sortAndPrint(ctx context.Context, results <-chan search.Result, formatter *output.Formatter) (int, error) {
	var all []search.Result
	for r := range results {
		if r.Err != nil {
			fmt.Fprintln(os.Stderr, formatter.FormatResult(r))
		} else {
			all = append(all, r)
		}
	}

	sort.Slice(all, func(i, j int) bool {
		if all[i].Path != all[j].Path {
			return all[i].Path < all[j].Path
		}
		return all[i].LineNum < all[j].LineNum
	})

	out := formatter.FormatGrouped(all)
	if out != "" {
		fmt.Print(out)
	}
	return len(all), nil
}
