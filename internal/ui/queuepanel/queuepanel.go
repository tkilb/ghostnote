package queuepanel

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/x/ansi"
	"github.com/tylerkilburn/ghostnote/internal/engine"
	"github.com/tylerkilburn/ghostnote/internal/ui/panel"
)

// Model represents the [2] Queue panel.
type Model struct {
	Queue   *engine.Queue
	Cursor  int
	Width   int
	Height  int
	Focused bool
}

// New creates a new queue panel model.
func New(q *engine.Queue) Model {
	return Model{
		Queue: q,
	}
}

// SetSize updates the panel's dimensions.
func (m *Model) SetSize(w, h int) {
	m.Width = w
	m.Height = h
}

// MoveUp moves cursor up.
func (m *Model) MoveUp() {
	if m.Cursor > 0 {
		m.Cursor--
	}
}

// MoveDown moves cursor down.
func (m *Model) MoveDown() {
	jobs := m.Queue.Jobs()
	if m.Cursor < len(jobs)-1 {
		m.Cursor++
	}
}

// SelectedJob returns the job under cursor, or nil.
func (m Model) SelectedJob() *engine.Job {
	jobs := m.Queue.Jobs()
	if len(jobs) == 0 || m.Cursor < 0 || m.Cursor >= len(jobs) {
		return nil
	}
	return jobs[m.Cursor]
}

// View renders the queue panel.
func (m Model) View() string {
	jobs := m.Queue.Jobs()
	title := fmt.Sprintf("[2] Queue (%d active)", len(jobs))

	if len(jobs) == 0 {
		body := " No items in queue. Press [a] to add a URL or playlist."
		return panel.Frame(title, body, m.Width, m.Height, m.Focused)
	}

	if m.Cursor >= len(jobs) {
		m.Cursor = len(jobs) - 1
	}
	if m.Cursor < 0 {
		m.Cursor = 0
	}

	var lines []string
	innerWidth := m.Width - 4
	if innerWidth < 10 {
		innerWidth = 10
	}

	for idx, j := range jobs {
		prefix := "  "
		if idx == m.Cursor {
			prefix = "> "
		}

		statusStr := fmt.Sprintf("[%s]", j.Status.String())
		if j.Progress > 0 && j.Status == engine.StatusDownloading {
			statusStr = fmt.Sprintf("[%.0f%% %s]", j.Progress, j.Speed)
		}

		itemText := fmt.Sprintf("%s%-30s %s", prefix, j.Title, statusStr)
		if idx == m.Cursor {
			itemText = panel.CursorRowStyle.Width(innerWidth).Render(itemText)
		} else {
			itemText = ansi.Truncate(itemText, innerWidth, "…")
		}
		lines = append(lines, itemText)
	}

	body := strings.Join(lines, "\n")
	return panel.Frame(title, body, m.Width, m.Height, m.Focused)
}
