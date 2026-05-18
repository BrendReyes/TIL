/*
Copyright © 2026 NAME HERE <EMAIL ADDRESS>

*/
package cmd

import (
	//"fmt"
	"strings"
	"github.com/spf13/cobra"

)



// addCmd represents the add command
var addCmd = &cobra.Command{
	Use:   "add <content>",
	Short: "Capture a new learning entry",
    Long: `Add a new entry to your TIL database.
   
Example:
  til add "In Go, defer runs LIFO" --tag go
  til add "BFS uses a queue, DFS uses a stack" -t algorithms`,

	Args: cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		content := strings.Join(args, " ")
		tag, _ := cmd.Flags().GetString("tag")
		return appState.AddEntry(content, tag)
	},
}

func init() {
	rootCmd.AddCommand(addCmd)
	addCmd.Flags().StringP("tag", "t", "", "Tag to categorize this entry (e.g. -t algorithms)")
	addCmd.MarkFlagRequired("tag")


	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// addCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// addCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
