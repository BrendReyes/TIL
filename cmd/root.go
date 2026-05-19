package cmd

import (
	"fmt"
	"os"
	"github.com/brendreyes/til/internal/srs"
	"github.com/spf13/cobra"
)

var appState *srs.State

func SetState(s *srs.State) {
	appState = s
}

const art = `
██╗ ████████╗    ██╗    ██╗                          
╚██╗╚══██╔══╝    ██║    ██║                          
 ╚██╗  ██║       ██║    ██║                          
 ██╔╝  ██║       ██║    ██║                          
██╔╝   ██║       ██║    ███████╗    ██╗    ██╗    ██╗
╚═╝    ╚═╝       ╚═╝    ╚══════╝    ╚═╝    ╚═╝    ╚═╝                                                                                              
`

var rootCmd = &cobra.Command{
	Use:   "til",
	Short: "A CLI tool for capturing and reviewing things you learn",
	Long: `TIL (Today I Learned) is a personal knowledge tracker designed to help you
retain information through spaced repetition.

Capture insights instantly from your terminal during study sessions and
review them later through an interactive TUI session to ensure they stick.`,
	Run: func(cmd *cobra.Command, args []string) {
        fmt.Printf(art)
        cmd.Help() 
   },
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.CompletionOptions.DisableDefaultCmd = true
	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}


