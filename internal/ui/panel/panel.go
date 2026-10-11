package panel

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
)

// Standard palette definitions matching lazytask / lazygit conventions:
// - Focused border: Green (ANSI 2)
// - Unfocused border: Default (terminal foreground / "")
// - Selection / Active: Blue (ANSI 4)
// - Error: Red (ANSI 1)
// - Warning / In-progress: Yellow (ANSI 3)
// - Info / Search: Cyan (ANSI 6)
// - Dim labels: 245
const (
	PaletteDefault = lipgloss.Color("")
	PaletteGreen   = lipgloss.Color("2")
	PaletteBlue    = lipgloss.Color("4")
	PaletteRed     = lipgloss.Color("1")
	PaletteYellow  = lipgloss.Color("3")
	PaletteCyan    = lipgloss.Color("6")
	PaletteDim     = lipgloss.Color("245")
)

var (
	UnfocusedColor = PaletteDefault
	FocusedColor   = PaletteGreen
	SelectedRowBg  = PaletteBlue
	ActiveTabColor = PaletteBlue
	KeyColor       = PaletteBlue

	BasePanelStyle = lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(UnfocusedColor)

	FocusedPanelStyle = lipgloss.NewStyle().
				BorderStyle(lipgloss.RoundedBorder()).
				BorderForeground(FocusedColor)

	PanelTitleFocused = lipgloss.NewStyle().
				Bold(true).
				Foreground(FocusedColor)

	PanelTitleUnfocused = lipgloss.NewStyle().
				Bold(true).
				Foreground(UnfocusedColor)

	CursorRowStyle = lipgloss.NewStyle().
			Background(SelectedRowBg).
			Bold(true)

	NormalRowStyle = lipgloss.NewStyle()

	StatusBarStyle = lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(UnfocusedColor)

	StatusKeyStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(KeyColor)

	StatusDescStyle = lipgloss.NewStyle().
			Foreground(PaletteDim)

	ModalBoxStyle = lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(FocusedColor).
			Padding(1, 2)
)

// BorderColor returns the border color based on focus.
func BorderColor(focused bool) lipgloss.Color {
	if focused {
		return FocusedColor
	}
	return UnfocusedColor
}

// TitleStyle returns the title style based on focus.
func TitleStyle(focused bool) lipgloss.Style {
	if focused {
		return PanelTitleFocused
	}
	return PanelTitleUnfocused
}

// Frame renders a lazygit-style panel with title embedded in the top border:
// e.g. "╭─[1] Config─────╮"
func Frame(title, body string, width, height int, focused bool, footer ...string) string {
	lineStyle := lipgloss.NewStyle().Foreground(BorderColor(focused))
	top := topBorderWithRight(title, "", width, focused, lineStyle)
	return frameBody(top, firstFooter(footer), body, width, height, lineStyle)
}

// FrameWithTopRight renders a panel with title on the left and a status tag on the top right.
func FrameWithTopRight(title, topRight, body string, width, height int, focused bool, footer ...string) string {
	lineStyle := lipgloss.NewStyle().Foreground(BorderColor(focused))
	top := topBorderWithRight(title, topRight, width, focused, lineStyle)
	return frameBody(top, firstFooter(footer), body, width, height, lineStyle)
}

func firstFooter(footer []string) string {
	if len(footer) == 0 {
		return ""
	}
	return footer[0]
}

func frameBody(top, footer, body string, width, height int, lineStyle lipgloss.Style) string {
	var innerLines []string
	if body != "" {
		innerLines = strings.Split(body, "\n")
	}

	contentHeight := height - 2
	if contentHeight < 0 {
		contentHeight = 0
	}

	for len(innerLines) < contentHeight {
		innerLines = append(innerLines, "")
	}
	if len(innerLines) > contentHeight {
		innerLines = innerLines[:contentHeight]
	}

	innerWidth := width - 2
	if innerWidth < 0 {
		innerWidth = 0
	}

	left := lineStyle.Render("│")
	right := lineStyle.Render("│")

	var rows []string
	rows = append(rows, top)
	for _, l := range innerLines {
		trunc := ansi.Truncate(l, innerWidth, "…")
		w := ansi.StringWidth(trunc)
		pad := innerWidth - w
		if pad < 0 {
			pad = 0
		}
		rows = append(rows, left+trunc+strings.Repeat(" ", pad)+right)
	}

	bot := bottomBorder(footer, width, lineStyle)
	rows = append(rows, bot)
	return strings.Join(rows, "\n")
}

func topBorderWithRight(title, rightText string, width int, focused bool, lineStyle lipgloss.Style) string {
	tl := lineStyle.Render("╭")
	tr := lineStyle.Render("╮")
	dash := lineStyle.Render("─")

	tStyle := TitleStyle(focused)
	rStyle := lipgloss.NewStyle().Foreground(PaletteDim)
	lowerRight := strings.ToLower(rightText)
	if strings.Contains(lowerRight, "not installed") || strings.Contains(lowerRight, "missing") {
		rStyle = lipgloss.NewStyle().Bold(true).Foreground(PaletteRed)
	}

	titleW := ansi.StringWidth(title)
	rightW := ansi.StringWidth(rightText)

	if titleW+rightW+6 > width {
		avail := width - rightW - 6
		if avail < 0 {
			avail = 0
		}
		title = ansi.Truncate(title, avail, "…")
		titleW = ansi.StringWidth(title)
	}

	rem := width - 2 - 1 - titleW
	if rightText != "" {
		rem -= (rightW + 1)
	}
	if rem < 0 {
		rem = 0
	}

	var sb strings.Builder
	sb.WriteString(tl)
	sb.WriteString(dash)
	sb.WriteString(tStyle.Render(title))

	if rightText != "" {
		sb.WriteString(strings.Repeat(dash, rem))
		sb.WriteString(rStyle.Render(rightText))
		sb.WriteString(dash)
	} else {
		sb.WriteString(strings.Repeat(dash, rem))
	}
	sb.WriteString(tr)
	return sb.String()
}

func bottomBorder(footer string, width int, lineStyle lipgloss.Style) string {
	bl := lineStyle.Render("╰")
	br := lineStyle.Render("╯")
	dash := lineStyle.Render("─")

	if footer == "" {
		rem := width - 2
		if rem < 0 {
			rem = 0
		}
		return bl + strings.Repeat(dash, rem) + br
	}

	fStyle := lipgloss.NewStyle().Foreground(PaletteDim)
	footerW := ansi.StringWidth(footer)
	if footerW+4 > width {
		avail := width - 4
		if avail < 0 {
			avail = 0
		}
		footer = ansi.Truncate(footer, avail, "…")
		footerW = ansi.StringWidth(footer)
	}

	rem := width - 2 - footerW - 2
	if rem < 0 {
		rem = 0
	}

	return bl + strings.Repeat(dash, rem) + dash + fStyle.Render(footer) + dash + br
}
