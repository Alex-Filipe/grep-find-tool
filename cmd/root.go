package cmd

import (
	"github.com/spf13/cobra"
)

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
	cobra.CheckErr(rootCmd.Execute())
}

func init() {
	rootCmd.PersistentFlags().IntVarP(&workers, "workers", "j", 4, "number of parallel workers")
}
