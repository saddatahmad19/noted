package cmd

import (
	"fmt"
	tui "cobra-cli/internal/tui"
	"github.com/spf13/cobra"
)

// vaultCmd represents the vault command
var vaultCmd = &cobra.Command{
	Use:   "vault",
	Short: "Open the vault selection/creation menu (TUI)",
	Long:  `Launches an interactive TUI to select or create a vault for Noted, using Bubble Tea and Lip Gloss.`,
	Run: func(cmd *cobra.Command, args []string) {
		// Get current vaults from config (reuse logic from root.go)
		// For now, just call the TUI with no vaults (for demo)
		vaults := []string{}
		selectedVault, err := tui.LaunchVaultTUI(vaults, "")
		if err != nil {
			fmt.Println("Error selecting vault:", err)
			return
		}
		fmt.Println("Vault set to:", selectedVault)
	},
}

func init() {
	rootCmd.AddCommand(vaultCmd)
}

// TODO: Implement Bubble Tea model, update, and view for the vault menu here or in a separate internal/tui package. 