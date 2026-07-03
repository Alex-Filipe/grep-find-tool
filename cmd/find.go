package cmd

import "github.com/spf13/cobra"

func init() {
	rootCmd.AddCommand(findCmd)
}

var findCmd = &cobra.Command{
	Use:   "find <name> [path...]",
	Short: "Search for files by name",
	Long: `Search for files matching the given name pattern.
Supports glob-style patterns like "*.go".`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return nil
	},
}
