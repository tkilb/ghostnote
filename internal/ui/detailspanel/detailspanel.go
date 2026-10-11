package detailspanel

import (
	"fmt"
	"strings"

	"github.com/tylerkilburn/ghostnote/internal/engine"
	"github.com/tylerkilburn/ghostnote/internal/ui/panel"
)

// Model represents the [0] Details panel.
type Model struct {
	Job     *engine.Job
	Width   int
	Height  int
	Focused bool
}

// New creates a new details panel model.
func New() Model {
	return Model{}
}

// SetJob updates the focused track/job.
func (m *Model) SetJob(job *engine.Job) {
	m.Job = job
}

// SetSize updates panel dimensions.
func (m *Model) SetSize(w, h int) {
	m.Width = w
	m.Height = h
}

// View renders track metadata and status details.
func (m Model) View() string {
	title := "[0] Details"
	if m.Job == nil {
		body := " Select a track in Queue or Completed to view details."
		return panel.Frame(title, body, m.Width, m.Height, m.Focused)
	}

	j := m.Job
	var lines []string
	lines = append(lines, fmt.Sprintf(" Title:    %s", j.Tags.Title))
	lines = append(lines, fmt.Sprintf(" Artist:   %s", j.Tags.Artist))
	lines = append(lines, fmt.Sprintf(" Album:    %s", j.Tags.Album))
	if j.Tags.TrackNumber > 0 {
		lines = append(lines, fmt.Sprintf(" Track:    %d/%d", j.Tags.TrackNumber, j.Tags.TotalTracks))
	} else {
		lines = append(lines, " Track:    -")
	}
	lines = append(lines, fmt.Sprintf(" Year:     %s", j.Tags.Year))
	lines = append(lines, fmt.Sprintf(" Format:   %s (%s)", strings.ToUpper(j.Format), j.Bitrate))
	lines = append(lines, fmt.Sprintf(" Source:   %s", j.URL))
	lines = append(lines, fmt.Sprintf(" Output:   %s", j.DestPath))
	lines = append(lines, fmt.Sprintf(" Status:   %s", j.Status.String()))

	if j.ErrorMsg != "" {
		lines = append(lines, fmt.Sprintf(" Error:    %s", j.ErrorMsg))
	}

	lines = append(lines, "")
	lines = append(lines, " [e] Edit Metadata")

	body := strings.Join(lines, "\n")
	return panel.Frame(title, body, m.Width, m.Height, m.Focused)
}
