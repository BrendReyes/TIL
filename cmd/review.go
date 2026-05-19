/*
Copyright © 2026 NAME HERE <EMAIL ADDRESS>

*/
package cmd

import (
	//"fmt"

	"github.com/spf13/cobra"
)

// reviewCmd represents the review command
var reviewCmd = &cobra.Command{
	Use:   "review",
	Short: "Start an interactive review session",
	Long: `Launch the interactive TUI to review entries that are due for a revisit.
It is just a simple recalling, no question and answers, but uses similar algorithm what Anki uses`,
	DisableFlagsInUseLine: true,
	Args: cobra.ExactArgs(0),
	RunE: func(cmd *cobra.Command, args []string) error {
		reset, _ := cmd.Flags().GetBool("reset")
        if reset {
            return appState.ResetAllReviews()
        }
        return appState.ReviewEntries()
	},
}

func init() {
	rootCmd.AddCommand(reviewCmd)
    reviewCmd.Flags().Bool("reset", false, "Reset all review progress so every entry becomes due immediately")
	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// reviewCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// reviewCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
