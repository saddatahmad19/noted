package models

// Vault represents a vault directory and its configuration.
// It embeds VaultConfig for direct access to config fields.
type Vault struct {
	Name           string // Human-readable vault name
	Path           string // Absolute path to the vault directory
	VaultConfigPath string // Path to the config file in the vault
	Config         VaultConfig // Embedded config for this vault
} 

