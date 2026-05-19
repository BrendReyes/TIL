package cmd

import (
	"github.com/spf13/cobra"
)

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
}
