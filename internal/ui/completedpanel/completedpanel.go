package completedpanel

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/x/ansi"
	"github.com/tylerkilburn/ghostnote/internal/engine"
	"github.com/tylerkilburn/ghostnote/internal/ui/panel"
)

// Model represents the [3] Completed panel.
type Model struct {
	Queue   *engine.Queue
	Cursor  int
	Width   int
	Height  int
	Focused bool
}

// New creates a new completed panel model.
func New(q *engine.Queue) Model {
	return Model{
		Queue: q,
	}
}

// SetSize updates the panel dimensions.
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
	items := m.Queue.Completed()
	if m.Cursor < len(items)-1 {
		m.Cursor++
	}
}

// SelectedJob returns the completed job under cursor.
func (m Model) SelectedJob() *engine.Job {
	items := m.Queue.Completed()
	if len(items) == 0 || m.Cursor < 0 || m.Cursor >= len(items) {
		return nil
	}
	return items[m.Cursor]
}

// View renders completed items.
func (m Model) View() string {
	items := m.Queue.Completed()
	title := fmt.Sprintf("[3] Completed (%d)", len(items))

	if len(items) == 0 {
		body := " No completed tracks yet."
		return panel.Frame(title, body, m.Width, m.Height, m.Focused)
	}

	if m.Cursor >= len(items) {
		m.Cursor = len(items) - 1
	}
	if m.Cursor < 0 {
		m.Cursor = 0
	}

	var lines []string
	innerWidth := m.Width - 4
	if innerWidth < 10 {
		innerWidth = 10
	}

	for idx, j := range items {
		prefix := "  "
		if idx == m.Cursor {
			prefix = "> "
		}

		itemText := fmt.Sprintf("%s%s", prefix, j.Title)
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
