package tui

import (
	"os"
	"path/filepath"
	"sort"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	textinput "github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/lipgloss"
)

var (
	dirPickerHeaderStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("63")).Padding(0, 1)
	dirPickerSelectedStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("230")).Background(lipgloss.Color("63")).Bold(true)
	dirPickerItemStyle = lipgloss.NewStyle().Padding(0, 1)
	dirPickerBorderStyle = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color("63")).Padding(1, 2)
	dirPickerErrorStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("9")).Bold(true)
)

type DirectoryPickerResult struct {
	Path      string
	Cancelled bool
	Err       error
}

type directoryPickerModel struct {
	input         textinput.Model
	dirs          []string
	filteredDirs  []string
	selectedIdx   int
	result        DirectoryPickerResult
	inputError    string
	state         int // 0: input, 1: done
}

func NewDirectoryPickerModel() directoryPickerModel {
	ti := textinput.New()
	ti.Placeholder = "~/Documents"
	ti.CharLimit = 256
	ti.Width = 48
	return directoryPickerModel{
		input: ti,
		dirs:  getInitialDirs(),
		filteredDirs: []string{},
		selectedIdx: 0,
		state: 0,
	}
}

func getInitialDirs() []string {
	home, err := os.UserHomeDir()
	if err != nil {
		home = "/"
	}
	return listDirs(home)
}

func listDirs(base string) []string {
	entries, err := os.ReadDir(base)
	if err != nil {
		return []string{base}
	}
	dirs := []string{}
	for _, entry := range entries {
		if entry.IsDir() {
			dirs = append(dirs, filepath.Join(base, entry.Name()))
		}
	}
	sort.Strings(dirs)
	return dirs
}

func (m directoryPickerModel) Init() tea.Cmd {
	return nil
}

func (m directoryPickerModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch m.state {
	case 0:
		switch msg := msg.(type) {
		case tea.KeyMsg:
			switch msg.String() {
			case "q", "esc":
				m.result.Cancelled = true
				m.state = 1
				return m, tea.Quit
			case "enter":
				if m.selectedIdx >= 0 && m.selectedIdx < len(m.filteredDirs) {
					m.result.Path = m.filteredDirs[m.selectedIdx]
					m.state = 1
					return m, tea.Quit
				}
				val := m.input.Value()
				if val == "" {
					m.inputError = "✗ Path cannot be empty."
					return m, nil
				}
				expanded, err := expandPath(val)
				if err != nil {
					m.inputError = "✗ Invalid path."
					return m, nil
				}
				m.result.Path = expanded
				m.state = 1
				return m, tea.Quit
			case "up":
				if m.selectedIdx > 0 {
					m.selectedIdx--
				}
				return m, nil
			case "down":
				if m.selectedIdx < len(m.filteredDirs)-1 {
					m.selectedIdx++
				}
				return m, nil
			}
		}
		var cmd tea.Cmd
		m.input, cmd = m.input.Update(msg)
		m.filteredDirs = m.filterDirs(m.input.Value())
		if m.selectedIdx >= len(m.filteredDirs) {
			m.selectedIdx = len(m.filteredDirs) - 1
		}
		if m.selectedIdx < 0 {
			m.selectedIdx = 0
		}
		return m, cmd
	}
	return m, nil
}

func (m directoryPickerModel) View() string {
	if m.state == 1 {
		return ""
	}
	prompt := dirPickerHeaderStyle.Render("Select or enter a directory:")
	inputBox := dirPickerBorderStyle.Render(m.input.View())
	dirList := ""
	for i, dir := range m.filteredDirs {
		name := dir
		if i == m.selectedIdx {
			dirList += dirPickerSelectedStyle.Render(name) + "\n"
		} else {
			dirList += dirPickerItemStyle.Render(name) + "\n"
		}
	}
	errMsg := ""
	if m.inputError != "" {
		errMsg = dirPickerErrorStyle.Render(m.inputError)
	}
	help := lipgloss.NewStyle().Foreground(lipgloss.Color("245")).Faint(true).Padding(0, 1).Render("[Enter] Select   [↑/↓] Navigate   [Esc] Cancel")
	return prompt + "\n" + inputBox + "\n" + dirList + errMsg + "\n" + help
}

func (m directoryPickerModel) filterDirs(prefix string) []string {
	if prefix == "" {
		return m.dirs
	}
	// Expand ~
	base, _ := expandPath(prefix)
	parent := filepath.Dir(base)
	if parent == "." || parent == "/" {
		parent, _ = os.UserHomeDir()
	}
	candidates := listDirs(parent)
	filtered := []string{}
	for _, d := range candidates {
		if strings.HasPrefix(d, base) {
			filtered = append(filtered, d)
		}
	}
	if len(filtered) == 0 && base != "" {
		filtered = append(filtered, base)
	}
	return filtered
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

// LaunchDirectoryPicker launches the directory picker and returns the selected path.
func LaunchDirectoryPicker() (string, error) {
	m := NewDirectoryPickerModel()
	p := tea.NewProgram(m)
	finalModel, err := p.Run()
	if err != nil {
		return "", err
	}
	result := finalModel.(directoryPickerModel).result
	if result.Cancelled {
		return "", os.ErrNotExist
	}
	return result.Path, result.Err
} 