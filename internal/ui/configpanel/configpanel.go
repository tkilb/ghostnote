package configpanel

import (
	"fmt"
	"strings"

	"github.com/tylerkilburn/ghostnote/internal/config"
	"github.com/tylerkilburn/ghostnote/internal/ui/panel"
)

// Model represents the [1] Config panel.
type Model struct {
	Cfg     *config.Config
	Width   int
	Height  int
	Focused bool
}

// New creates a new config panel model.
func New(cfg *config.Config) Model {
	return Model{
		Cfg: cfg,
	}
}

// SetSize updates the panel's dimensions.
func (m *Model) SetSize(w, h int) {
	m.Width = w
	m.Height = h
}

// View renders the panel content.
func (m Model) View() string {
	dest := m.Cfg.Destination
	format := strings.ToUpper(m.Cfg.Audio.Format)
	bitrate := m.Cfg.Audio.Bitrate
	pattern := m.Cfg.Naming.Pattern
	ostPattern := m.Cfg.Naming.OSTPattern

	var lines []string
	lines = append(lines, fmt.Sprintf(" Dest:    %s", dest))
	lines = append(lines, fmt.Sprintf(" Format:  %s (%s)", format, bitrate))
	lines = append(lines, fmt.Sprintf(" Pattern: %s", pattern))
	lines = append(lines, fmt.Sprintf(" OST Pat: %s", ostPattern))
	lines = append(lines, "")
	lines = append(lines, " [d] change destination  [f] toggle format")

	body := strings.Join(lines, "\n")
	return panel.Frame("[1] Config", body, m.Width, m.Height, m.Focused)
}
