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
)

var configDir string
var configFile string
var notedConfig *viper.Viper

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "noted",
	Short: "A CLI tool for managing notes with vaults, templates, and fast search.",
	Long: `Noted is a CLI tool inspired by Obsidian, designed for thorough note management.
It supports vaults, templates, and fast search, with configuration stored in $XDG_CONFIG_HOME/noted or ~/.config/noted.`,
	Run: func(cmd *cobra.Command, args []string) {
		initConfigDir()
		initConfigFile()
		ensureVault()
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
	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	// rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.cobra-cli.yaml)")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
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

// ensureVault checks if a vault is set, prompts user if not
func ensureVault() {
	vaults := notedConfig.GetStringSlice("vaults")
	currentVault := notedConfig.GetString("current_vault")
	if currentVault == "" || !contains(vaults, currentVault) {
		// Launch Bubble Tea TUI for vault selection/creation
		selectedVault, err := tui.LaunchVaultTUI(vaults, currentVault)
		if err != nil {
			fmt.Println("Error selecting vault:", err)
			os.Exit(1)
		}
		if !contains(vaults, selectedVault) {
			vaults = append(vaults, selectedVault)
		}
		notedConfig.Set("current_vault", selectedVault)
		notedConfig.Set("vaults", vaults)
		err = notedConfig.WriteConfigAs(configFile)
		if err != nil {
			fmt.Println("Failed to update config:", err)
			os.Exit(1)
		}
		fmt.Println("Vault set to:", selectedVault)
		return
	}
	fmt.Println("Current vault:", currentVault)
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


