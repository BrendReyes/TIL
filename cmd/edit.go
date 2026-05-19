package cmd

import (
	"fmt"
	"strconv"
	"github.com/spf13/cobra"
)

var editCmd = &cobra.Command{
	Use:   "edit <id>",
	Short: "Modify an existing entry's content",
	Long: `Open a specific entry in the interactive TUI editor.

Use this to fix typos, update tags, or expand on your notes as your
understanding of the topic evolves.`,
	DisableFlagsInUseLine: true,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		id, err := strconv.ParseInt(args[0], 10, 64)
        if err != nil {
            return fmt.Errorf("invalid ID %q — must be a number", args[0])
        }
		
		return appState.EditEntry(id)
	},
}

func init() {
	rootCmd.AddCommand(editCmd)
}
