package cmd

import (
	"github.com/spf13/cobra"
)

// helpCmd represents the custom help command
var helpCmd = &cobra.Command{
	Use:   "help",
	Short: "Show a beautiful TUI help screen",
	Long:  `Displays a styled, interactive help screen using Bubble Tea and Lip Gloss.`,
	Run: func(cmd *cobra.Command, args []string) {
		// TODO: Launch Bubble Tea TUI for help
		// This will use the model and view logic for a scrollable, styled help screen
	},
}

func init() {
	rootCmd.SetHelpCommand(helpCmd)
}

// TODO: Implement Bubble Tea model, update, and view for the help screen here or in a separate internal/tui package. 