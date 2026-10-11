package editmodal

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/tylerkilburn/ghostnote/internal/engine"
	"github.com/tylerkilburn/ghostnote/internal/ui/panel"
)

type FieldIndex int

const (
	FieldTitle FieldIndex = iota
	FieldArtist
	FieldAlbum
	FieldTrack
	FieldYear
	totalFields
)

// Model represents the inline metadata editing form modal.
type Model struct {
	Job     *engine.Job
	Inputs  []textinput.Model
	FocusIdx int
	Active  bool
	Width   int
	Height  int
}

// New creates a new metadata edit modal.
func New() Model {
	inputs := make([]textinput.Model, totalFields)

	fields := []struct {
		placeholder string
		charLimit   int
	}{
		{"Title", 120},
		{"Artist", 100},
		{"Album", 100},
		{"Track #", 10},
		{"Year", 10},
	}

	for i := range inputs {
		ti := textinput.New()
		ti.Placeholder = fields[i].placeholder
		ti.CharLimit = fields[i].charLimit
		ti.Width = 36
		inputs[i] = ti
	}

	return Model{
		Inputs: inputs,
	}
}

// Open activates the modal and pre-fills values from the given job.
func (m *Model) Open(job *engine.Job) {
	if job == nil {
		return
	}
	m.Job = job
	m.Active = true
	m.FocusIdx = 0

	m.Inputs[FieldTitle].SetValue(job.Tags.Title)
	m.Inputs[FieldArtist].SetValue(job.Tags.Artist)
	m.Inputs[FieldAlbum].SetValue(job.Tags.Album)
	if job.Tags.TrackNumber > 0 {
		m.Inputs[FieldTrack].SetValue(strconv.Itoa(job.Tags.TrackNumber))
	} else {
		m.Inputs[FieldTrack].SetValue("")
	}
	m.Inputs[FieldYear].SetValue(job.Tags.Year)

	for i := range m.Inputs {
		if i == 0 {
			m.Inputs[i].Focus()
		} else {
			m.Inputs[i].Blur()
		}
	}
}

// Close deactivates the modal.
func (m *Model) Close() {
	m.Active = false
	for i := range m.Inputs {
		m.Inputs[i].Blur()
	}
}

// Save applies form fields back to the Job's Tags and regenerates output path if needed.
func (m *Model) Save() {
	if m.Job == nil {
		return
	}

	m.Job.Tags.Title = strings.TrimSpace(m.Inputs[FieldTitle].Value())
	m.Job.Title = m.Job.Tags.Title
	m.Job.Tags.Artist = strings.TrimSpace(m.Inputs[FieldArtist].Value())
	m.Job.Tags.Album = strings.TrimSpace(m.Inputs[FieldAlbum].Value())
	m.Job.Tags.Year = strings.TrimSpace(m.Inputs[FieldYear].Value())

	trackStr := strings.TrimSpace(m.Inputs[FieldTrack].Value())
	if num, err := strconv.Atoi(trackStr); err == nil {
		m.Job.Tags.TrackNumber = num
	}
}

// Update handles tab navigation and typing inside the inputs.
func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	if !m.Active {
		return m, nil
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "tab", "down":
			m.focusNext()
			return m, nil
		case "shift+tab", "up":
			m.focusPrev()
			return m, nil
		}
	}

	var cmd tea.Cmd
	m.Inputs[m.FocusIdx], cmd = m.Inputs[m.FocusIdx].Update(msg)
	return m, cmd
}

func (m *Model) focusNext() {
	m.Inputs[m.FocusIdx].Blur()
	m.FocusIdx = (m.FocusIdx + 1) % int(totalFields)
	m.Inputs[m.FocusIdx].Focus()
}

func (m *Model) focusPrev() {
	m.Inputs[m.FocusIdx].Blur()
	m.FocusIdx = (m.FocusIdx + int(totalFields) - 1) % int(totalFields)
	m.Inputs[m.FocusIdx].Focus()
}

// View renders the metadata editor modal.
func (m Model) View(width, height int) string {
	if !m.Active {
		return ""
	}

	boxWidth := 56
	if width < 60 {
		boxWidth = width - 4
	}

	title := panel.PanelTitleFocused.Render("Edit Metadata Tags")
	var lines []string
	lines = append(lines, title)
	lines = append(lines, "")

	labels := []string{"Title:   ", "Artist:  ", "Album:   ", "Track #: ", "Year:    "}
	for i, l := range labels {
		lines = append(lines, fmt.Sprintf(" %s %s", l, m.Inputs[i].View()))
	}

	lines = append(lines, "")
	lines = append(lines, panel.StatusDescStyle.Render(" [Tab/↑/↓] Switch Field   [Enter] Save   [Esc] Cancel"))

	content := strings.Join(lines, "\n")
	box := panel.ModalBoxStyle.Width(boxWidth).Render(content)

	return lipgloss.Place(width, height, lipgloss.Center, lipgloss.Center, box)
}
