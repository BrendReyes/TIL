package cmd

import (
	"fmt"
	"strconv"
	"strings"
	"github.com/spf13/cobra"
)

var deleteCmd = &cobra.Command{
	Use:   `delete all or delete <id>`,
	Short: "Delete an entry",
	Long: `Permanently remove entry from the storage

Example
  til delete all
  til delete 5
  til delete all -t python
  til delete all --tag "system design"
	`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if strings.ToLower(args[0]) == "all" {
			tag, _ := cmd.Flags().GetString("tag")
			if tag != "" {
				return appState.RemoveEntryByTag(tag)
			}
			return appState.RemoveAllEntry()
		}

		id, err := strconv.ParseInt(args[0], 10, 64)
        if err != nil {
            return fmt.Errorf("invalid ID %q — must be a valid id", args[0])
        }
		
        return appState.DeleteEntry(id)
	},
}

func init() {
	rootCmd.AddCommand(deleteCmd)
	deleteCmd.Flags().StringP("tag", "t", "", "Delete all entries by tag (e.g til delete all -t sql)")
}
