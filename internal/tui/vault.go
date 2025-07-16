package tui

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	tea "github.com/charmbracelet/bubbletea"
	list "github.com/charmbracelet/bubbles/list"
	textinput "github.com/charmbracelet/bubbles/textinput"

	"github.com/charmbracelet/lipgloss"
	log "github.com/charmbracelet/log"

	"cobra-cli/internal/models"
)

// --- Lip Gloss Styles ---
var (
	headerStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("63")).Padding(0, 1)
	selectedStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("230")).Background(lipgloss.Color("63")).Bold(true)
	itemStyle = lipgloss.NewStyle().Padding(0, 1)
	buttonStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("99")).Bold(true)
	helpBarStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("245")).Faint(true).Padding(0, 1)
	errorStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("9")).Bold(true)
	successStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("10")).Bold(true)
	borderStyle = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color("63")).Padding(1, 2)
	modalStyle = lipgloss.NewStyle().Border(lipgloss.NormalBorder()).BorderForeground(lipgloss.Color("99")).Padding(1, 2).Background(lipgloss.Color("236"))
	infoPanelStyle = lipgloss.NewStyle().Border(lipgloss.NormalBorder()).BorderForeground(lipgloss.Color("245")).Padding(1, 2).Margin(1, 0)
)

// VaultTUIResult holds the result of the TUI
// Name: the name of the vault
// Path: the selected or created vault path
// Cancelled: true if the user cancelled
// Err: any error
//
type VaultTUIResult struct {
	Name      string
	Path      string
	Cancelled bool
	Err       error
}

// LaunchVaultTUI launches the Bubble Tea TUI for vault selection/creation
func LaunchVaultTUI(vaults []models.Vault, currentVault string) (models.Vault, error) {
	m := newVaultModel(vaults)
	p := tea.NewProgram(m)
	finalModel, err := p.Run()
	if err != nil {
		return models.Vault{}, err
	}
	result := finalModel.(vaultModel).result
	if result.Cancelled {
		return models.Vault{}, fmt.Errorf("vault selection cancelled")
	}
	if result.Err != nil {
		return models.Vault{}, result.Err
	}
	return models.Vault{Name: result.Name, Path: result.Path}, nil
}

// --- Bubble Tea Model ---

type vaultState int

const (
	stateList vaultState = iota
	stateInput
	stateNameInput
	stateConfirmCreate
	stateDone
	stateDirPicker
	stateDeleteConfirm
)

type vaultMainMenu int

const (
	menuList vaultMainMenu = iota
	menuDetails
	menuEdit
	menuCreate
	menuDelete
	menuExit
)

type vaultModel struct {
	vaults      []models.Vault
	list        list.Model
	input       textinput.Model
	nameInput   textinput.Model
	state       vaultState
	menu        vaultMainMenu
	selectedIdx int
	result      VaultTUIResult
	inputPrompt string
	inputError  string
	confirmPath string
	editField   int
	// Add fields for editing config
	editingConfig models.VaultConfig

	dirPicker   directoryPickerModel // Directory picker component
	showDirPicker bool
	showDeleteConfirm bool
	deleteIdx   int
}

func newVaultModel(vaults []models.Vault) vaultModel {
	items := make([]list.Item, len(vaults)+1)
	for i, v := range vaults {
		items[i] = vaultListItem{v.Name, v.Path}
	}
	items[len(vaults)] = vaultListItem{"+ Create New Vault", ""}
	l := list.New(items, list.NewDefaultDelegate(), 40, 10)
	l.Title = "Select a Vault for Noted"
	l.SetShowStatusBar(false)
	l.SetShowHelp(false)
	l.SetShowPagination(false)
	ti := textinput.New()
	ti.Placeholder = "~/Documents/PersonalKnowledge"
	ti.CharLimit = 256
	ti.Width = 36
	ni := textinput.New()
	ni.Placeholder = "Vault Name"
	ni.CharLimit = 64
	ni.Width = 36
	return vaultModel{
		vaults:    vaults,
		list:      l,
		input:     ti,
		nameInput: ni,
		state:     stateList,
		dirPicker: NewDirectoryPickerModel(),
	}
}

type vaultListItem struct {
	name string
	path string
}

func (i vaultListItem) Title() string       { return i.name }
func (i vaultListItem) Description() string { return i.path }
func (i vaultListItem) FilterValue() string { return i.name + " " + i.path }

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
					m.state = stateDirPicker
					m.dirPicker = NewDirectoryPickerModel()
					return m, nil
				}
				m.result.Name = m.vaults[idx].Name
				m.result.Path = m.vaults[idx].Path
				m.state = stateDone
				return m, tea.Quit
			case "D":
				idx := m.list.Index()
				if idx < len(m.vaults) {
					m.state = stateDeleteConfirm
					m.deleteIdx = idx
					return m, nil
				}
			}
		}
		var cmd tea.Cmd
		m.list, cmd = m.list.Update(msg)
		return m, cmd
	case stateDirPicker:
		model, cmd := m.dirPicker.Update(msg)
		m.dirPicker = model.(directoryPickerModel)
		if m.dirPicker.state == 1 {
			if m.dirPicker.result.Cancelled {
				m.state = stateList
				return m, nil
			}
			m.result.Path = m.dirPicker.result.Path
			m.state = stateNameInput
			m.nameInput.SetValue(filepath.Base(m.result.Path))
			m.nameInput.Focus()
			return m, nil
		}
		return m, cmd
	case stateNameInput:
		switch msg := msg.(type) {
		case tea.KeyMsg:
			switch msg.String() {
			case "esc":
				m.nameInput.Blur() // Blur name input when leaving
				m.state = stateInput
				m.input.Focus() // Refocus path input
				return m, nil
			case "enter":
				name := m.nameInput.Value()
				if name == "" {
					m.inputError = "✗ Name cannot be empty."
					return m, nil
				}
				m.nameInput.Blur() // Blur name input on done
				m.result.Name = name
				// Write VaultConfig to the selected directory
				cfg := models.VaultConfig{
					Name: name,
					TemplatesPath: filepath.Join(m.result.Path, "templates"),
					LogPath: filepath.Join(m.result.Path, "vault.log"),
					HistoryPath: filepath.Join(m.result.Path, "history.log"),
					SupportedTypes: []string{".md", ".pdf"},
					IgnorePatterns: []string{".git", "node_modules"},
					Metadata: map[string]string{},
					Settings: map[string]any{},
				}
				if err := writeVaultConfig(m.result.Path, cfg); err != nil {
					log.Error("Failed to write vault config", "err", err)
					m.result.Err = err
					m.state = stateDone
					return m, tea.Quit
				}
				log.Info("Vault created", "path", m.result.Path)
				m.state = stateDone
				return m, tea.Quit
			}
		}
		var cmd tea.Cmd
		m.nameInput, cmd = m.nameInput.Update(msg)
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
				m.state = stateNameInput
				m.nameInput.SetValue(filepath.Base(m.confirmPath))
				m.nameInput.Focus() // Focus name input after creating dir
				return m, nil
			case "n", "N", "esc":
				m.state = stateInput
				m.input.Focus() // Refocus path input
				return m, nil
			}
		}
		return m, nil
	case stateDeleteConfirm:
		switch msg := msg.(type) {
		case tea.KeyMsg:
			switch msg.String() {
			case "y", "Y":
				idx := m.deleteIdx
				if idx < len(m.vaults) {
					vault := m.vaults[idx]
					// Remove vault config file and directory (optional: prompt for full delete)
					cfgPath := filepath.Join(vault.Path, "vault.json")
					os.Remove(cfgPath)
					log.Info("Vault deleted", "path", vault.Path)
					m.vaults = append(m.vaults[:idx], m.vaults[idx+1:]...)
					items := make([]list.Item, len(m.vaults)+1)
					for i, v := range m.vaults {
						items[i] = vaultListItem{v.Name, v.Path}
					}
					items[len(m.vaults)] = vaultListItem{"+ Create New Vault", ""}
					m.list.SetItems(items)
					m.state = stateList
					return m, nil
				}
			case "n", "N", "esc":
				m.state = stateList
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
		// Styled vault list
		var items []string
		for i, item := range m.list.Items() {
			str := item.(vaultListItem).name
			if i == m.list.Index() {
				if str == "+ Create New Vault" {
					str = buttonStyle.Render(str)
				} else {
					str = selectedStyle.Render(str)
				}
			} else {
				if str == "+ Create New Vault" {
					str = buttonStyle.Render(str)
				} else {
					str = itemStyle.Render(str)
				}
			}
			items = append(items, str)
		}
		listView := headerStyle.Render("Select a Vault for Noted") + "\n" + lipgloss.JoinVertical(lipgloss.Left, items...)
		help := helpBarStyle.Render("↑/↓: Move   Enter: Select   q: Quit")
		return borderStyle.Render(listView+"\n\n"+help)
	case stateInput:
		prompt := headerStyle.Render(m.inputPrompt)
		inputBox := borderStyle.Render(m.input.View())
		errMsg := ""
		if m.inputError != "" {
			errMsg = errorStyle.Render(m.inputError)
		}
		help := helpBarStyle.Render("[Enter] Confirm   [Esc] Cancel")
		return prompt + "\n\n" + inputBox + "\n" + errMsg + "\n" + help
	case stateNameInput:
		prompt := headerStyle.Render("Enter a name for your new vault (default: folder name):")
		inputBox := borderStyle.Render(m.nameInput.View())
		errMsg := ""
		if m.inputError != "" {
			errMsg = errorStyle.Render(m.inputError)
		}
		help := helpBarStyle.Render("[Enter] Confirm   [Esc] Back")
		return prompt + "\n\n" + inputBox + "\n" + errMsg + "\n" + help
	case stateConfirmCreate:
		modal := modalStyle.Render("Vault directory does not exist.\nCreate it? [y/N]")
		return "\n" + modal
	case stateDirPicker:
		return m.dirPicker.View()
	case stateDeleteConfirm:
		vault := m.vaults[m.deleteIdx]
		modal := modalStyle.Render(fmt.Sprintf("Delete vault '%s'? [y/N]", vault.Name))
		return "\n" + modal
	case stateDone:
		if m.result.Err != nil {
			return errorStyle.Render("Error: "+m.result.Err.Error())
		}
		if m.result.Cancelled {
			return errorStyle.Render("Cancelled.")
		}
		msg := successStyle.Render("✓ Vault set!") + "\n" + headerStyle.Render("Current vault: "+m.result.Name) + "\n" + itemStyle.Render(m.result.Path) + "\n" + helpBarStyle.Render("[Enter] Continue")
		return borderStyle.Render(msg)
	}
	return ""
}

// Add a helper to render the details panel for a vault
func renderVaultDetails(v models.Vault) string {
	cfg := v.Config
	return infoPanelStyle.Render(
		headerStyle.Render("Vault Details") + "\n" +
		itemStyle.Render("Name: ") + cfg.Name + "\n" +
		itemStyle.Render("Path: ") + v.Path + "\n" +
		itemStyle.Render("Templates: ") + cfg.TemplatesPath + "\n" +
		itemStyle.Render("Log: ") + cfg.LogPath + "\n" +
		itemStyle.Render("History: ") + cfg.HistoryPath + "\n" +
		itemStyle.Render("Supported Types: ") + fmt.Sprintf("%v", cfg.SupportedTypes) + "\n" +
		itemStyle.Render("Ignore Patterns: ") + fmt.Sprintf("%v", cfg.IgnorePatterns) + "\n" +
		itemStyle.Render("Created: ") + cfg.CreatedAt.Format("2006-01-02 15:04:05") + "\n" +
		itemStyle.Render("Modified: ") + cfg.ModifiedAt.Format("2006-01-02 15:04:05") + "\n" +
		itemStyle.Render("Metadata: ") + fmt.Sprintf("%v", cfg.Metadata) + "\n" +
		itemStyle.Render("Settings: ") + fmt.Sprintf("%v", cfg.Settings) + "\n",
	)
}

// Helper to write VaultConfig to disk
func writeVaultConfig(path string, cfg models.VaultConfig) error {
	cfgPath := filepath.Join(path, "vault.json")
	f, err := os.Create(cfgPath)
	if err != nil {
		return err
	}
	defer f.Close()
	// Use encoding/json to marshal config
	enc := json.NewEncoder(f)
	enc.SetIndent("", "  ")
	return enc.Encode(cfg)
} 