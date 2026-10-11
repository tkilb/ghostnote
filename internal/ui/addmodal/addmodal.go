package addmodal

import (
	"strings"

	"github.com/atotto/clipboard"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/tylerkilburn/ghostnote/internal/ui/panel"
)

// Model represents the URL input modal.
type Model struct {
	TextInput textinput.Model
	Active    bool
	Width     int
	Height    int
}

// New creates an add modal.
func New() Model {
	ti := textinput.New()
	ti.Placeholder = "https://www.youtube.com/watch?v=..."
	ti.Focus()
	ti.CharLimit = 500
	ti.Width = 50

	return Model{
		TextInput: ti,
	}
}

// Open activates the modal and pre-fills from system clipboard if valid URL.
func (m *Model) Open() {
	m.Active = true
	m.TextInput.Reset()
	if clip, err := clipboard.ReadAll(); err == nil {
		trimmed := strings.TrimSpace(clip)
		if strings.HasPrefix(trimmed, "http://") || strings.HasPrefix(trimmed, "https://") {
			m.TextInput.SetValue(trimmed)
		}
	}
	m.TextInput.Focus()
}

// Close deactivates the modal.
func (m *Model) Close() {
	m.Active = false
	m.TextInput.Blur()
}

// Update handles textinput keys.
func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	if !m.Active {
		return m, nil
	}

	var cmd tea.Cmd
	m.TextInput, cmd = m.TextInput.Update(msg)
	return m, cmd
}

// Value returns current input text.
func (m Model) Value() string {
	return strings.TrimSpace(m.TextInput.Value())
}

// View renders the modal centered over current terminal dimensions.
func (m Model) View(width, height int) string {
	if !m.Active {
		return ""
	}

	boxWidth := 58
	if width < 62 {
		boxWidth = width - 4
	}

	title := panel.PanelTitleFocused.Render("Add URL or Playlist")
	var lines []string
	lines = append(lines, title)
	lines = append(lines, "")
	lines = append(lines, " Enter media URL (YouTube, SoundCloud, Bandcamp, direct stream):")
	lines = append(lines, "")
	lines = append(lines, " "+m.TextInput.View())
	lines = append(lines, "")
	lines = append(lines, panel.StatusDescStyle.Render(" [Enter] Queue Track/Mix   [Esc] Cancel"))

	content := strings.Join(lines, "\n")
	box := panel.ModalBoxStyle.Width(boxWidth).Render(content)

	return lipgloss.Place(width, height, lipgloss.Center, lipgloss.Center, box)
}
