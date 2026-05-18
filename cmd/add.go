/*
Copyright © 2026 NAME HERE <EMAIL ADDRESS>

*/
package cmd

import (
	//"fmt"
	
	"github.com/spf13/cobra"

)



// addCmd represents the add command
var addCmd = &cobra.Command{
	Use:   "add",
	Short: "Add what you have learned",
	Long: `Add what you have learned:
		   til add "BFS uses a queue, DFS uses a stack"
		   til add BFS uses a queue, DFS uses a stack
		  `,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		tag, _ := cmd.Flags().GetString("tag")
		return appState.AddEntry(args[0], tag)
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
