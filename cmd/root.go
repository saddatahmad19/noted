/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>

*/
package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	tui "cobra-cli/internal/tui"
	"cobra-cli/internal/models"
	"encoding/json"
)

var configDir string
var configFile string
var notedConfig *viper.Viper

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "noted",
	Short: "A CLI tool for managing notes with vaults, templates, and fast search.",
	Long: `Noted is a CLI tool inspired by Obsidian, designed for thorough note management.
It supports vaults, templates, and fast search, with configuration stored in $XDG_CONFIG_HOME/noted or ~/.config/noted.

Run 'noted' without arguments to see the tutorial menu with all available commands.`,
	Run: func(cmd *cobra.Command, args []string) {
		initConfigDir()
		initConfigFile()
		
		// Show tutorial menu
		showTutorialMenu()
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfigDir, initConfigFile)
	
	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

func showTutorialMenu() {
	currentVault := notedConfig.GetString("current_vault")
	vaults := notedConfig.GetStringSlice("vaults")
	
	fmt.Println("╔═══════════════════════════════════════════════════════════════════════════════╗")
	fmt.Println("║                                   📝 NOTED                                    ║")
	fmt.Println("║                    A CLI tool for managing notes and vaults                   ║")
	fmt.Println("╚═══════════════════════════════════════════════════════════════════════════════╝")
	fmt.Println()
	
	// Show current vault status
	if currentVault != "" {
		fmt.Printf("📂 Current vault: %s\n", getVaultName(currentVault))
		fmt.Printf("   Path: %s\n", currentVault)
	} else {
		fmt.Println("⚠️  No vault selected. Use 'noted vault' to select or create one.")
	}
	fmt.Println()
	
	// Show available commands
	fmt.Println("🔧 Available Commands:")
	fmt.Println()
	
	fmt.Println("  📁 VAULT MANAGEMENT:")
	fmt.Println("    noted vault                    # Interactive vault selection/creation menu")
	fmt.Println("    noted vault --open <name>      # Open vault by name")
	fmt.Println("    noted vault --open <index>     # Open vault by index (1-based)")
	fmt.Println("    noted vault list               # List all configured vaults")
	fmt.Println("    noted vault current            # Show current vault")
	fmt.Println("    noted vault create <path>      # Create new vault at specified path")
	fmt.Println()
	
	fmt.Println("  🔍 SEARCH & NAVIGATION:")
	fmt.Println("    noted search                   # Search through files and directories")
	fmt.Println("    noted search --files           # Search files only")
	fmt.Println("    noted search --dirs            # Search directories only")
	fmt.Println()
	
	fmt.Println("  📋 TEMPLATES:")
	fmt.Println("    noted templates                # Manage and browse templates")
	fmt.Println("    noted templates list           # List available templates")
	fmt.Println("    noted templates create         # Create new file from template")
	fmt.Println()
	
	fmt.Println("  ℹ️  HELP & INFO:")
	fmt.Println("    noted help                     # Show this help menu")
	fmt.Println("    noted version                  # Show version information")
	fmt.Println()
	
	// Show quick start guide
	if len(vaults) == 0 {
		fmt.Println("🚀 Quick Start:")
		fmt.Println("   1. Run 'noted vault' to create your first vault")
		fmt.Println("   2. Select a directory to use as your notes vault")
		fmt.Println("   3. Start organizing your notes!")
		fmt.Println()
	} else {
		fmt.Println("🚀 Quick Actions:")
		fmt.Println("   • 'noted vault' - Switch or create vaults")
		fmt.Println("   • 'noted search' - Find files and notes")
		fmt.Println("   • 'noted templates' - Use templates for new files")
		fmt.Println()
	}
	
	fmt.Println("💡 Tip: Most commands support interactive menus for easy navigation!")
	fmt.Println()
	
	// Show vault list if any exist
	if len(vaults) > 0 {
		fmt.Println("📋 Your Vaults:")
		for i, vault := range vaults {
			current := ""
			if vault == currentVault {
				current = " (current)"
			}
			fmt.Printf("   %d. %s%s\n", i+1, getVaultName(vault), current)
		}
		fmt.Println()
	}
	
	fmt.Println("Run any command to get started, or 'noted help' for more information.")
}

// initConfigDir checks for $XDG_CONFIG_HOME/noted or ~/.config/noted, creates if missing
func initConfigDir() {
	xdgConfig := os.Getenv("XDG_CONFIG_HOME")
	if xdgConfig == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			fmt.Println("Could not determine home directory:", err)
			os.Exit(1)
		}
		xdgConfig = filepath.Join(home, ".config")
	}
	configDir = filepath.Join(xdgConfig, "noted")
	if _, err := os.Stat(configDir); os.IsNotExist(err) {
		err := os.MkdirAll(configDir, 0o755)
		if err != nil {
			fmt.Println("Failed to create config directory:", err)
			os.Exit(1)
		}
		fmt.Println("Created config directory at", configDir)
	}
}

// initConfigFile initializes the config file using Viper
func initConfigFile() {
	configFile = filepath.Join(configDir, "config.yaml")
	notedConfig = viper.New()
	notedConfig.SetConfigFile(configFile)
	notedConfig.SetConfigType("yaml")

	if _, err := os.Stat(configFile); os.IsNotExist(err) {
		// Create default config
		notedConfig.Set("vaults", []string{})
		notedConfig.Set("current_vault", "")
		notedConfig.Set("templates_dir", "")
		notedConfig.Set("other_settings", map[string]interface{}{})
		err := notedConfig.WriteConfigAs(configFile)
		if err != nil {
			fmt.Println("Failed to write default config:", err)
			os.Exit(1)
		}
		fmt.Println("Initialized new config at", configFile)
	} else {
		err := notedConfig.ReadInConfig()
		if err != nil {
			fmt.Println("Failed to read config:", err)
			os.Exit(1)
		}
	}
}

// Helper to load vaults from config
func loadVaults() []models.Vault {
	var vaults []models.Vault
	vaultsRaw := notedConfig.Get("vaults")
	if vaultsRaw == nil {
		return vaults
	}
	// Try to unmarshal from JSON if stored as string
	switch v := vaultsRaw.(type) {
	case string:
		_ = json.Unmarshal([]byte(v), &vaults)
	case []interface{}:
		for _, item := range v {
			if m, ok := item.(map[string]interface{}); ok {
				vaults = append(vaults, models.Vault{
					Name: m["Name"].(string),
					Path: m["Path"].(string),
				})
			}
		}
	}
	return vaults
}

func saveVaults(vaults []models.Vault) {
	b, _ := json.Marshal(vaults)
	notedConfig.Set("vaults", string(b))
}

func vaultContains(vaults []models.Vault, v models.Vault) bool {
	for _, vault := range vaults {
		if vault.Path == v.Path {
			return true
		}
	}
	return false
}

func ensureVault() {
	vaults := loadVaults()
	currentVault := notedConfig.GetString("current_vault")
	found := false
	for _, v := range vaults {
		if v.Path == currentVault {
			found = true
			break
		}
	}
	if currentVault == "" || !found {
		selectedVault, err := tui.LaunchVaultTUI(vaults, currentVault)
		if err != nil {
			fmt.Println("Error selecting vault:", err)
			os.Exit(1)
		}
		if !vaultContains(vaults, selectedVault) {
			vaults = append(vaults, selectedVault)
			saveVaults(vaults)
		}
		notedConfig.Set("current_vault", selectedVault.Path)
		err = notedConfig.WriteConfigAs(configFile)
		if err != nil {
			fmt.Println("Failed to update config:", err)
			os.Exit(1)
		}
		fmt.Println("Vault set to:", selectedVault.Name)
		return
	}
	for _, v := range vaults {
		if v.Path == currentVault {
			fmt.Println("Current vault:", v.Name)
			return
		}
	}
	fmt.Println("Current vault path:", currentVault, "(not found in vaults list)")
}

func contains(slice []string, s string) bool {
	for _, v := range slice {
		if v == s {
			return true
		}
	}
	return false
}

func parseIndex(input string, max int) (int, error) {
	idx := -1
	_, err := fmt.Sscanf(input, "%d", &idx)
	if err != nil || idx < 1 || idx > max {
		return 0, fmt.Errorf("invalid index")
	}
	return idx - 1, nil
}

func getVaultName(path string) string {
	return filepath.Base(path)
}