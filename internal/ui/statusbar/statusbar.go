package statusbar

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/tylerkilburn/ghostnote/internal/ui/panel"
)

// View renders the bottom key hints bar.
func View(width int) string {
	hints := []struct {
		key  string
		desc string
	}{
		{"[1-3,0/Tab]", "panels"},
		{"[↑/↓]", "nav"},
		{"[a]", "add"},
		{"[d]", "dest"},
		{"[f]", "format"},
		{"[?]", "help"},
		{"[q]", "quit"},
	}

	var parts []string
	for _, h := range hints {
		k := panel.StatusKeyStyle.Render(h.key)
		d := panel.StatusDescStyle.Render(h.desc)
		parts = append(parts, k+" "+d)
	}

	text := " " + strings.Join(parts, "  ")
	return lipgloss.NewStyle().Width(width).Render(text)
}
