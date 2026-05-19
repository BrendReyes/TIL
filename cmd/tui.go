package cmd

import (
	"github.com/spf13/cobra"
)

var tuiCmd = &cobra.Command{
	Use:     "tui",
	Aliases: []string{"t"},
	Short:   "Start the interactive TUI",
	Long:    `Launch the interactive Terminal User Interface to manage your learning entries.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		//here calls the tui
		//return appState.RunMainTUI()
		return nil
	},
}

func init() {
	rootCmd.AddCommand(tuiCmd)
}
