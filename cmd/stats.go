/*
Copyright © 2026 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"github.com/spf13/cobra"
)

var statsCmd = &cobra.Command{
	Use:   "stats",
	Short: "Show statistics about your learning entries",
	Long: `Display a summary of your TIL database.

Shows total entries, reviewed vs unreviewed, due today, and a breakdown by tag.

Example:
  til stats`,
	DisableFlagsInUseLine: true,
	Args:                  cobra.ExactArgs(0),
	RunE: func(cmd *cobra.Command, args []string) error {
		return appState.ShowStats()
	},
}

func init() {
	rootCmd.AddCommand(statsCmd)
}