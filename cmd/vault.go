package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	tui "cobra-cli/internal/tui"
	"cobra-cli/internal/models"
	"github.com/spf13/cobra"
)

var openFlag string

// vaultCmd represents the vault command
var vaultCmd = &cobra.Command{
	Use:   "vault",
	Short: "Manage vaults - select, create, or open a specific vault",
	Long: `Manage your note vaults with the following options:
	
  noted vault                    # Open interactive vault selection/creation menu
  noted vault --open <name>      # Open vault by name
  noted vault --open <index>     # Open vault by index (1-based)
  noted vault list               # List all configured vaults
  noted vault current            # Show current vault
  noted vault create <path>      # Create new vault at specified path`,
	Run: func(cmd *cobra.Command, args []string) {
		initConfigDir()
		initConfigFile()
		
		// Handle --open flag
		if openFlag != "" {
			openVaultByNameOrIndex(openFlag)
			return
		}
		
		// No flags, launch interactive TUI
		launchVaultTUI()
	},
}

var vaultListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all configured vaults",
	Long:  `Display a list of all configured vaults with their indices and paths.`,
	Run: func(cmd *cobra.Command, args []string) {
		initConfigDir()
		initConfigFile()
		listVaults()
	},
}

var vaultCurrentCmd = &cobra.Command{
	Use:   "current",
	Short: "Show current active vault",
	Long:  `Display the currently active vault path and name.`,
	Run: func(cmd *cobra.Command, args []string) {
		initConfigDir()
		initConfigFile()
		showCurrentVault()
	},
}

var vaultCreateCmd = &cobra.Command{
	Use:   "create <path>",
	Short: "Create a new vault at the specified path",
	Long:  `Create a new vault at the specified path and set it as the current vault.`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		initConfigDir()
		initConfigFile()
		createVault(args[0])
	},
}

func init() {
	rootCmd.AddCommand(vaultCmd)
	vaultCmd.AddCommand(vaultListCmd)
	vaultCmd.AddCommand(vaultCurrentCmd)
	vaultCmd.AddCommand(vaultCreateCmd)
	
	// Add --open flag
	vaultCmd.Flags().StringVarP(&openFlag, "open", "o", "", "Open vault by name or index")
}

func launchVaultTUI() {
	vaults := loadVaults()
	currentVault := notedConfig.GetString("current_vault")
	selectedVault, err := tui.LaunchVaultTUI(vaults, currentVault)
	if err != nil {
		fmt.Printf("Error selecting vault: %v\n", err)
		return
	}
	// Update config
	if !vaultContains(vaults, selectedVault) {
		vaults = append(vaults, selectedVault)
		saveVaults(vaults)
	}
	notedConfig.Set("current_vault", selectedVault.Path)
	err = notedConfig.WriteConfigAs(configFile)
	if err != nil {
		fmt.Printf("Failed to update config: %v\n", err)
		return
	}
	fmt.Printf("✓ Vault set to: %s\n", selectedVault.Name)
	launchVaultViewer(selectedVault.Path)
}

func openVaultByNameOrIndex(input string) {
	vaults := loadVaults()
	if len(vaults) == 0 {
		fmt.Println("No vaults configured. Run 'noted vault' to create one.")
		return
	}
	var selectedVault *models.Vault
	if idx, err := strconv.Atoi(input); err == nil {
		if idx < 1 || idx > len(vaults) {
			fmt.Printf("Invalid vault index: %d. Valid range: 1-%d\n", idx, len(vaults))
			return
		}
		selectedVault = &vaults[idx-1]
	} else {
		for i := range vaults {
			if vaults[i].Name == input {
				selectedVault = &vaults[i]
				break
			}
		}
		if selectedVault == nil {
			fmt.Printf("Vault '%s' not found.\n", input)
			fmt.Println("Available vaults:")
			for i, vault := range vaults {
				fmt.Printf("  %d. %s (%s)\n", i+1, vault.Name, vault.Path)
			}
			return
		}
	}
	notedConfig.Set("current_vault", selectedVault.Path)
	err := notedConfig.WriteConfigAs(configFile)
	if err != nil {
		fmt.Printf("Failed to update config: %v\n", err)
		return
	}
	fmt.Printf("✓ Opened vault: %s\n", selectedVault.Name)
	launchVaultViewer(selectedVault.Path)
}

func listVaults() {
	vaults := loadVaults()
	currentVault := notedConfig.GetString("current_vault")
	if len(vaults) == 0 {
		fmt.Println("No vaults configured. Run 'noted vault' to create one.")
		return
	}
	fmt.Println("Configured vaults:")
	for i, vault := range vaults {
		current := ""
		if vault.Path == currentVault {
			current = " (current)"
		}
		fmt.Printf("  %d. %s%s\n     %s\n", i+1, vault.Name, current, vault.Path)
	}
}

func showCurrentVault() {
	currentVault := notedConfig.GetString("current_vault")
	vaults := loadVaults()
	if currentVault == "" {
		fmt.Println("No current vault set. Run 'noted vault' to select one.")
		return
	}
	for _, vault := range vaults {
		if vault.Path == currentVault {
			fmt.Printf("Current vault: %s\n", vault.Name)
			fmt.Printf("Path: %s\n", vault.Path)
			return
		}
	}
	fmt.Printf("Current vault path: %s (not found in vaults list)\n", currentVault)
}

func createVault(path string) {
	expanded, err := expandPath(path)
	if err != nil {
		fmt.Printf("Error expanding path: %v\n", err)
		return
	}
	if _, err := os.Stat(expanded); os.IsNotExist(err) {
		fmt.Printf("Directory '%s' does not exist. Create it? [y/N]: ", expanded)
		var response string
		fmt.Scanln(&response)
		if strings.ToLower(response) != "y" && strings.ToLower(response) != "yes" {
			fmt.Println("Vault creation cancelled.")
			return
		}
		err := os.MkdirAll(expanded, 0o755)
		if err != nil {
			fmt.Printf("Failed to create directory: %v\n", err)
			return
		}
		fmt.Printf("✓ Created directory: %s\n", expanded)
	}
	vaults := loadVaults()
	name := filepath.Base(expanded)
	newVault := models.Vault{Name: name, Path: expanded}
	if !vaultContains(vaults, newVault) {
		vaults = append(vaults, newVault)
		saveVaults(vaults)
	}
	notedConfig.Set("current_vault", newVault.Path)
	err = notedConfig.WriteConfigAs(configFile)
	if err != nil {
		fmt.Printf("Failed to update config: %v\n", err)
		return
	}
	fmt.Printf("✓ Vault created and set as current: %s\n", newVault.Name)
	launchVaultViewer(newVault.Path)
}

func launchVaultViewer(vaultPath string) {
	fmt.Printf("\nLaunching vault viewer for: %s\n", filepath.Base(vaultPath))
	// TODO: Implement vault viewer TUI
	// err := tui.LaunchVaultViewer(vaultPath)
	// if err != nil {
	// 	fmt.Printf("Error launching vault viewer: %v\n", err)
	// 	return
	// }
}

// expandPath expands ~ to home directory
func expandPath(path string) (string, error) {
	if len(path) > 0 && path[0] == '~' {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		return filepath.Join(home, path[1:]), nil
	}
	return path, nil
}