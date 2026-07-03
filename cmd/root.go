package cmd

import "github.com/spf13/cobra"

var (
	workers    int
	ignoreCase bool
)

var rootCmd = &cobra.Command{
	Use:   "grep-tool",
	Short: "A fast text search tool for files",
}

func Execute() {
	cobra.CheckErr(rootCmd.Execute())
}

func init() {
	rootCmd.PersistentFlags().IntVarP(&workers, "workers", "j", 4, "number of parallel workers")
	rootCmd.PersistentFlags().BoolVarP(&ignoreCase, "ignore-case", "i", false, "case-insensitive search")
}
