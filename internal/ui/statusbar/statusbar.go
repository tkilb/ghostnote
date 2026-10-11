package statusbar

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/tylerkilburn/ghostnote/internal/ui/panel"
)

// View renders the bottom key hints bar and optional warning tag.
func View(width int, warning string) string {
	hints := []struct {
		key  string
		desc string
	}{
		{"[1-3,0/Tab]", "panels"},
		{"[↑/↓]", "nav"},
		{"[a]", "add"},
		{"[d]", "dest"},
		{"[f]", "format"},
		{"[r]", "retry"},
		{"[?]", "help"},
		{"[q]", "quit"},
	}

	var parts []string
	for _, h := range hints {
		k := panel.StatusKeyStyle.Render(h.key)
		d := panel.StatusDescStyle.Render(h.desc)
		parts = append(parts, k+" "+d)
	}

	leftText := " " + strings.Join(parts, "  ")
	if warning == "" {
		return lipgloss.NewStyle().Width(width).Render(leftText)
	}

	warnTag := lipgloss.NewStyle().Bold(true).Foreground(panel.PaletteRed).Render(" ⚠ " + warning + " ")
	return lipgloss.JoinHorizontal(lipgloss.Top,
		leftText,
		strings.Repeat(" ", max(0, width-lipgloss.Width(leftText)-lipgloss.Width(warnTag))),
		warnTag,
	)
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
