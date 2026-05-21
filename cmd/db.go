package cmd

import (
    "github.com/spf13/cobra"
)

var dbCmd = &cobra.Command{
    Use:   "db",
    Short: "Database management commands",
}

var dbPathCmd = &cobra.Command{
    Use:   "path",
    Short: "Print the path to the database file",
    Run: func(cmd *cobra.Command, args []string) {
        appState.ShowPath()
    },
}

func init() {
    rootCmd.AddCommand(dbCmd)
    dbCmd.AddCommand(dbPathCmd)
}