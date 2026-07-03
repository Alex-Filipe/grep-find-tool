package cmd

import "github.com/spf13/cobra"

func init() {
	rootCmd.AddCommand(grepCmd)
}

var grepCmd = &cobra.Command{
	Use:   "grep <pattern> [path...]",
	Short: "Search for text inside files",
	Long: `Search recursively for files containing the given pattern.
Supports literal and regex matching with case-insensitive option.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return nil
	},
}
