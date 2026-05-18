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
	Short: "List all of the entries",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Args: cobra.ExactArgs(0),
	RunE: func(cmd *cobra.Command, args []string) error {
		idFlag, _ := cmd.Flags().GetString("id")
 
		if idFlag != "" {
			id, err := strconv.ParseInt(idFlag, 10, 64)
			if err != nil {
				return fmt.Errorf("invalid ID %q — must be a number", idFlag)
			}
			return appState.GetSpecificEntry(id)
		}
 
		return appState.ListEntry()

	},
}

func init() {
	rootCmd.AddCommand(listCmd)
	listCmd.Flags().StringP("id", "i", "", "View a specific entry by ID (e.g. til list -i 5)")

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// listCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// listCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
