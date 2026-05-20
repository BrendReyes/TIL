package cmd

import (
	"strings"

	"github.com/spf13/cobra"
)

var addCmd = &cobra.Command{
	Use:   "add [content]",
	Short: "Capture a new learning entry",
	Long: `Add a new entry to your TIL database.

If no content is provided, opens the interactive TUI editor.
If content is provided as arguments, saves directly from the CLI.

Examples:
  til add                                            (opens TUI)
  til add "In Go, defer runs LIFO" --tag go
  til add BFS uses a queue, DFS uses a stack -t algorithms`,

	Args: cobra.ArbitraryArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		// No args — open the TUI add screen
		if len(args) == 0 {
			return appState.RunAddTUI()
		}

		// Args provided — CLI path
		content := strings.Join(args, " ")
		tag, _ := cmd.Flags().GetString("tag")
		return appState.AddEntry(content, tag)
	},
}

func init() {
	rootCmd.AddCommand(addCmd)
	addCmd.Flags().StringP("tag", "t", "", "Tag to categorize this entry (e.g. -t algorithms)")
}