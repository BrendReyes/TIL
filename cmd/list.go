/*
Copyright © 2026 NAME HERE <EMAIL ADDRESS>

*/
package cmd

import (
	"fmt"
	"strconv"
	"github.com/spf13/cobra"
)

// listCmd represents the list command
var listCmd = &cobra.Command{
	Use:   "list",
	Short: "Display entries or view a specific one",
	Long: `Display a summarized table of all captured learning entries.
   
Examples:
  til list (displays all)
  til list --tag postgres
  til list -t algorithms
  til list --id 5
  til list -i 3`,
	Args: cobra.ExactArgs(0),
	RunE: func(cmd *cobra.Command, args []string) error {
		idFlag, _ := cmd.Flags().GetString("id")
		tagFlag, _ := cmd.Flags().GetString("tag")

		if idFlag != "" {
			id, err := strconv.ParseInt(idFlag, 10, 64)
			if err != nil {
				return fmt.Errorf("invalid ID %q — must be a number", idFlag)
			}
			return appState.GetSpecificEntry(id)
		}

		if tagFlag != "" {
			return appState.ListEntriesByTag(tagFlag)
		}
 
		return appState.ListEntry()

	},
}

func init() {
	rootCmd.AddCommand(listCmd)
	listCmd.Flags().StringP("id", "i", "", "View a specific entry by ID (e.g. til list -i 5)")
	listCmd.Flags().StringP("tag", "t", "", "List all entries by tags (e.g. til list -t postgres)")

	listCmd.MarkFlagsMutuallyExclusive("id", "tag")
}
