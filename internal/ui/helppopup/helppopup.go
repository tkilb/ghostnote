package helppopup

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/tylerkilburn/ghostnote/internal/ui/panel"
)

// View renders the help modal overlay.
func View(width, height int) string {
	boxWidth := 56
	if width < 60 {
		boxWidth = width - 4
	}

	title := panel.PanelTitleFocused.Render("Ghostnote Keybindings")
	var lines []string
	lines = append(lines, title)
	lines = append(lines, "")
	lines = append(lines, "  0, 1, 2, 3   Jump to specific panel")
	lines = append(lines, "  Tab / S-Tab  Cycle through panels")
	lines = append(lines, "  ↑ / ↓        Navigate items in active panel")
	lines = append(lines, "  a            Add URL to queue")
	lines = append(lines, "  d            Pick destination directory")
	lines = append(lines, "  f            Toggle target audio format")
	lines = append(lines, "  x            Remove / cancel selected job")
	lines = append(lines, "  c            Clear completed jobs")
	lines = append(lines, "  e            Edit metadata of selected track")
	lines = append(lines, "  ?            Toggle this help modal")
	lines = append(lines, "  q / Ctrl+C   Quit Ghostnote")
	lines = append(lines, "")
	lines = append(lines, panel.StatusDescStyle.Render(" Press [Esc] or [?] to close"))

	content := strings.Join(lines, "\n")
	box := panel.ModalBoxStyle.Width(boxWidth).Render(content)

	return lipgloss.Place(width, height, lipgloss.Center, lipgloss.Center, box)
}
