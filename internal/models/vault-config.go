package models

import "time"

// VaultConfig represents the configuration for a vault.
// This config is stored at the base of the vault directory.
type VaultConfig struct {
	Name           string            `json:"name"`             // Human-readable vault name
	TemplatesPath  string            `json:"templates_path"`   // Path to the templates directory
	LogPath        string            `json:"log_path"`         // Path to the vault's log file (optional)
	HistoryPath    string            `json:"history_path"`     // Path to the vault's history file (optional)
	SupportedTypes []string          `json:"supported_types"`  // File extensions/types supported (e.g., [".md", ".pdf"])
	IgnorePatterns []string          `json:"ignore_patterns"`  // Glob patterns to ignore (e.g., [".git", "node_modules"])
	CreatedAt      time.Time         `json:"created_at"`       // When the vault was created
	ModifiedAt     time.Time         `json:"modified_at"`      // Last modified time
	Metadata       map[string]string `json:"metadata"`         // Arbitrary metadata (tags, etc.)
	Settings       map[string]any    `json:"settings"`         // Arbitrary custom settings for extensibility
}
