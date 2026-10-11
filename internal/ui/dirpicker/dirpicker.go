package dirpicker

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/tylerkilburn/ghostnote/internal/config"
	"github.com/tylerkilburn/ghostnote/internal/ui/panel"
)

// Model represents the directory navigation modal.
type Model struct {
	CurrentPath string
	Entries     []string
	Cursor      int
	Active      bool
	Width       int
	Height      int
}

// New creates a new directory picker model.
func New(initialDir string) Model {
	dir := config.ExpandHome(initialDir)
	if dir == "" {
		home, _ := os.UserHomeDir()
		dir = home
	}

	m := Model{
		CurrentPath: dir,
	}
	m.readEntries()
	return m
}

// Open activates the picker.
func (m *Model) Open(currentDir string) {
	m.Active = true
	m.CurrentPath = config.ExpandHome(currentDir)
	m.Cursor = 0
	m.readEntries()
}

// Close deactivates the picker.
func (m *Model) Close() {
	m.Active = false
}

func (m *Model) readEntries() {
	m.Entries = []string{".."}
	entries, err := os.ReadDir(m.CurrentPath)
	if err != nil {
		return
	}

	var dirs []string
	for _, e := range entries {
		if e.IsDir() && !strings.HasPrefix(e.Name(), ".") {
			dirs = append(dirs, e.Name())
		}
	}
	sort.Strings(dirs)
	m.Entries = append(m.Entries, dirs...)
}

// MoveUp moves cursor up.
func (m *Model) MoveUp() {
	if m.Cursor > 0 {
		m.Cursor--
	}
}

// MoveDown moves cursor down.
func (m *Model) MoveDown() {
	if m.Cursor < len(m.Entries)-1 {
		m.Cursor++
	}
}

// SelectEntry descends into directory or ..
func (m *Model) SelectEntry() {
	if len(m.Entries) == 0 {
		return
	}
	selected := m.Entries[m.Cursor]
	if selected == ".." {
		m.CurrentPath = filepath.Dir(m.CurrentPath)
	} else {
		m.CurrentPath = filepath.Join(m.CurrentPath, selected)
	}
	m.Cursor = 0
	m.readEntries()
}

// Update handles key navigation.
func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	if !m.Active {
		return m, nil
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "up":
			m.MoveUp()
		case "down":
			m.MoveDown()
		case "enter":
			m.SelectEntry()
		}
	}

	return m, nil
}

// View renders the directory picker modal.
func (m Model) View(width, height int) string {
	if !m.Active {
		return ""
	}

	boxWidth := 60
	if width < 64 {
		boxWidth = width - 4
	}

	title := panel.PanelTitleFocused.Render("Select Destination Directory")
	var lines []string
	lines = append(lines, title)
	lines = append(lines, fmt.Sprintf(" Path: %s", m.CurrentPath))
	lines = append(lines, "")

	maxShow := 10
	start := 0
	if m.Cursor >= maxShow {
		start = m.Cursor - maxShow + 1
	}
	end := start + maxShow
	if end > len(m.Entries) {
		end = len(m.Entries)
	}

	for i := start; i < end; i++ {
		e := m.Entries[i]
		line := fmt.Sprintf("   📁 %s", e)
		if i == m.Cursor {
			line = panel.CursorRowStyle.Render(fmt.Sprintf(" > 📁 %s", e))
		}
		lines = append(lines, line)
	}

	lines = append(lines, "")
	lines = append(lines, panel.StatusDescStyle.Render(" [Enter] Open Folder  [Space] Confirm Path  [Esc] Cancel"))

	content := strings.Join(lines, "\n")
	box := panel.ModalBoxStyle.Width(boxWidth).Render(content)

	return lipgloss.Place(width, height, lipgloss.Center, lipgloss.Center, box)
}
