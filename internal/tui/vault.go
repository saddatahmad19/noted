package tui

import (
	"fmt"
	"os"
	"path/filepath"

	tea "github.com/charmbracelet/bubbletea"
	list "github.com/charmbracelet/bubbles/list"
	textinput "github.com/charmbracelet/bubbles/textinput"
)

// VaultTUIResult holds the result of the TUI
// Path: the selected or created vault path
// Cancelled: true if the user cancelled
// Err: any error
//
type VaultTUIResult struct {
	Path      string
	Cancelled bool
	Err       error
}

// LaunchVaultTUI launches the Bubble Tea TUI for vault selection/creation
func LaunchVaultTUI(vaults []string, currentVault string) (string, error) {
	m := newVaultModel(vaults)
	p := tea.NewProgram(m)
	finalModel, err := p.Run()
	if err != nil {
		return "", err
	}
	result := finalModel.(vaultModel).result
	if result.Cancelled {
		return "", fmt.Errorf("vault selection cancelled")
	}
	if result.Err != nil {
		return "", result.Err
	}
	return result.Path, nil
}

// --- Bubble Tea Model ---

type vaultState int

const (
	stateList vaultState = iota
	stateInput
	stateConfirmCreate
	stateDone
)

type vaultModel struct {
	vaults      []string
	list        list.Model
	input       textinput.Model
	state       vaultState
	result      VaultTUIResult
	inputPrompt string
	inputError  string
	confirmPath string
}

func newVaultModel(vaults []string) vaultModel {
	items := make([]list.Item, len(vaults)+1)
	for i, v := range vaults {
		items[i] = listItem(v)
	}
	items[len(vaults)] = listItem("+ Create New Vault")
	l := list.New(items, list.NewDefaultDelegate(), 40, 10)
	l.Title = "Select a Vault for Noted"
	l.SetShowStatusBar(false)
	l.SetShowHelp(false)
	l.SetShowPagination(false)
	ti := textinput.New()
	ti.Placeholder = "~/Documents/PersonalKnowledge"
	ti.CharLimit = 256
	ti.Width = 36

	return vaultModel{
		vaults: vaults,
		list:   l,
		input:  ti,
		state:  stateList,
	}
}

type listItem string

func (i listItem) Title() string       { return string(i) }
func (i listItem) Description() string { return "" }
func (i listItem) FilterValue() string { return string(i) }

func (m vaultModel) Init() tea.Cmd {
	return nil
}

func (m vaultModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch m.state {
	case stateList:
		switch msg := msg.(type) {
		case tea.KeyMsg:
			switch msg.String() {
			case "q", "esc":
				m.result.Cancelled = true
				m.state = stateDone
				return m, tea.Quit
			case "enter":
				idx := m.list.Index()
				if idx == len(m.vaults) {
					// Create new vault
					m.state = stateInput
					m.inputPrompt = "Enter the full path for your new vault:"
					m.input.SetValue("")
					m.inputError = ""
					return m, nil
				}
				// Select existing vault
				m.result.Path = m.vaults[idx]
				m.state = stateDone
				return m, tea.Quit
			}
		}
		var cmd tea.Cmd
		m.list, cmd = m.list.Update(msg)
		return m, cmd
	case stateInput:
		switch msg := msg.(type) {
		case tea.KeyMsg:
			switch msg.String() {
			case "esc":
				m.state = stateList
				return m, nil
			case "enter":
				path := m.input.Value()
				if path == "" {
					m.inputError = "✗ Path cannot be empty."
					return m, nil
				}
				expanded, err := expandPath(path)
				if err != nil {
					m.inputError = "✗ Invalid path."
					return m, nil
				}
				if _, err := os.Stat(expanded); os.IsNotExist(err) {
					m.state = stateConfirmCreate
					m.confirmPath = expanded
					return m, nil
				}
				// Path exists
				m.result.Path = expanded
				m.state = stateDone
				return m, tea.Quit
			}
		}
		var cmd tea.Cmd
		m.input, cmd = m.input.Update(msg)
		return m, cmd
	case stateConfirmCreate:
		switch msg := msg.(type) {
		case tea.KeyMsg:
			switch msg.String() {
			case "y", "Y":
				if err := os.MkdirAll(m.confirmPath, 0o755); err != nil {
					m.result.Err = fmt.Errorf("failed to create directory: %w", err)
					m.state = stateDone
					return m, tea.Quit
				}
				m.result.Path = m.confirmPath
				m.state = stateDone
				return m, tea.Quit
			case "n", "N", "esc":
				m.state = stateInput
				return m, nil
			}
		}
		return m, nil
	}
	return m, nil
}

func (m vaultModel) View() string {
	switch m.state {
	case stateList:
		return m.list.View() + "\n↑/↓: Move   Enter: Select   q: Quit"
	case stateInput:
		return m.inputPrompt + "\n\n  " + m.input.View() + "\n" + m.inputError + "\n[Enter] Confirm   [Esc] Cancel"
	case stateConfirmCreate:
		return "Vault directory does not exist.\nCreate it? [y/N]"
	case stateDone:
		if m.result.Err != nil {
			return "Error: " + m.result.Err.Error()
		}
		if m.result.Cancelled {
			return "Cancelled."
		}
		return "Vault set!\nCurrent vault: " + m.result.Path + "\n[Enter] Continue"
	}
	return ""
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