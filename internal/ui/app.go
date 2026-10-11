package ui

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/tylerkilburn/ghostnote/internal/config"
	"github.com/tylerkilburn/ghostnote/internal/engine"
	"github.com/tylerkilburn/ghostnote/internal/ui/addmodal"
	"github.com/tylerkilburn/ghostnote/internal/ui/completedpanel"
	"github.com/tylerkilburn/ghostnote/internal/ui/configpanel"
	"github.com/tylerkilburn/ghostnote/internal/ui/detailspanel"
	"github.com/tylerkilburn/ghostnote/internal/ui/dirpicker"
	"github.com/tylerkilburn/ghostnote/internal/ui/editmodal"
	"github.com/tylerkilburn/ghostnote/internal/ui/helppopup"
	"github.com/tylerkilburn/ghostnote/internal/ui/queuepanel"
	"github.com/tylerkilburn/ghostnote/internal/ui/statusbar"
)

type FocusedPanel int

const (
	PanelDetails FocusedPanel = 0
	PanelConfig   FocusedPanel = 1
	PanelQueue    FocusedPanel = 2
	PanelCompleted FocusedPanel = 3
)

// AppModel represents root Bubble Tea application state.
type AppModel struct {
	Cfg          *config.Config
	Queue        *engine.Queue
	Focused      FocusedPanel
	ConfigPanel  configpanel.Model
	QueuePanel   queuepanel.Model
	CompPanel    completedpanel.Model
	DetailsPanel detailspanel.Model
	AddModal     addmodal.Model
	DirPicker    dirpicker.Model
	EditModal    editmodal.Model
	ShowHelp     bool

	DependencyWarning string

	Width  int
	Height int
}

// JobUpdateMsg notifies UI of job changes.
type JobUpdateMsg struct {
	Job *engine.Job
}

type queueBridge struct {
	p *tea.Program
}

func (b *queueBridge) OnJobUpdated(job *engine.Job) {
	if b.p != nil {
		b.p.Send(JobUpdateMsg{Job: job})
	}
}

// NewAppModel creates and initializes the application model.
func NewAppModel(cfg *config.Config, q *engine.Queue) AppModel {
	var depWarn string
	if err := engine.CheckYtDlp(); err != nil {
		depWarn = "yt-dlp is not installed"
	} else if err := engine.CheckFFmpeg(); err != nil {
		depWarn = "ffmpeg is not installed"
	}

	cp := configpanel.New(cfg)
	cp.Warning = depWarn

	return AppModel{
		Cfg:               cfg,
		Queue:             q,
		Focused:           PanelQueue,
		ConfigPanel:       cp,
		QueuePanel:        queuepanel.New(q),
		CompPanel:         completedpanel.New(q),
		DetailsPanel:      detailspanel.New(),
		AddModal:          addmodal.New(),
		DirPicker:         dirpicker.New(cfg.Destination),
		EditModal:         editmodal.New(),
		DependencyWarning: depWarn,
	}
}

func (m AppModel) listenQueueEvents() tea.Cmd {
	return func() tea.Msg {
		if m.Queue == nil {
			return nil
		}
		job := <-m.Queue.Events()
		return JobUpdateMsg{Job: job}
	}
}

// Init starts the Bubble Tea program.
func (m AppModel) Init() tea.Cmd {
	return m.listenQueueEvents()
}

// Update handles incoming messages and keybindings.
func (m AppModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.Width = msg.Width
		m.Height = msg.Height
		m.updateSizes()
		return m, nil

	case JobUpdateMsg:
		m.updateDetailsSelection()
		return m, m.listenQueueEvents()

	case tea.KeyMsg:
		// Always honor Ctrl+C globally, regardless of modal state or active inputs
		if msg.Type == tea.KeyCtrlC || msg.String() == "ctrl+c" {
			return m, tea.Quit
		}

		// Modal priority
		if m.ShowHelp {
			if msg.String() == "esc" || msg.String() == "?" || msg.String() == "q" {
				m.ShowHelp = false
			}
			return m, nil
		}

		if m.AddModal.Active {
			switch msg.String() {
			case "esc":
				m.AddModal.Close()
				return m, nil
			case "enter":
				url := m.AddModal.Value()
				if url != "" {
					_ = m.Queue.AddURLWithOptions(url, m.AddModal.AutoSplit)
				}
				m.AddModal.Close()
				m.updateDetailsSelection()
				return m, nil
			default:
				var cmd tea.Cmd
				m.AddModal, cmd = m.AddModal.Update(msg)
				return m, cmd
			}
		}

		if m.DirPicker.Active {
			switch msg.String() {
			case "esc":
				m.DirPicker.Close()
				return m, nil
			case " ": // Confirm path
				m.Cfg.Destination = m.DirPicker.CurrentPath
				_ = m.Cfg.Save()
				m.DirPicker.Close()
				return m, nil
			default:
				var cmd tea.Cmd
				m.DirPicker, cmd = m.DirPicker.Update(msg)
				return m, cmd
			}
		}

		if m.EditModal.Active {
			switch msg.String() {
			case "esc":
				m.EditModal.Close()
				return m, nil
			case "enter":
				m.EditModal.Save()
				m.EditModal.Close()
				m.updateDetailsSelection()
				return m, nil
			default:
				var cmd tea.Cmd
				m.EditModal, cmd = m.EditModal.Update(msg)
				return m, cmd
			}
		}

		// Global keys
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "?":
			m.ShowHelp = true
			return m, nil
		case "0":
			m.setFocus(PanelDetails)
			return m, nil
		case "1":
			m.setFocus(PanelConfig)
			return m, nil
		case "2":
			m.setFocus(PanelQueue)
			return m, nil
		case "3":
			m.setFocus(PanelCompleted)
			return m, nil
		case "tab":
			m.nextFocus()
			return m, nil
		case "shift+tab":
			m.prevFocus()
			return m, nil
		case "a":
			cmd := m.AddModal.Open()
			return m, cmd
		case "d":
			m.DirPicker.Open(m.Cfg.Destination)
			return m, nil
		case "f":
			m.toggleFormat()
			return m, nil
		case "e":
			if m.DetailsPanel.Job != nil {
				m.EditModal.Open(m.DetailsPanel.Job)
			}
			return m, nil
		case "up":
			m.handleUp()
			m.updateDetailsSelection()
			return m, nil
		case "down":
			m.handleDown()
			m.updateDetailsSelection()
			return m, nil
		case "x":
			if m.Focused == PanelQueue {
				if sel := m.QueuePanel.SelectedJob(); sel != nil {
					m.Queue.RemoveJob(sel.ID)
					m.updateDetailsSelection()
				}
			}
			return m, nil
		case "r":
			if m.Focused == PanelQueue {
				if sel := m.QueuePanel.SelectedJob(); sel != nil {
					m.Queue.RetryJob(sel.ID)
					m.updateDetailsSelection()
				}
			}
			return m, nil
		case "c":
			m.Queue.ClearCompleted()
			m.updateDetailsSelection()
			return m, nil
		}
	}

	return m, nil
}

func (m *AppModel) setFocus(p FocusedPanel) {
	m.Focused = p
	m.ConfigPanel.Focused = (p == PanelConfig)
	m.QueuePanel.Focused = (p == PanelQueue)
	m.CompPanel.Focused = (p == PanelCompleted)
	m.DetailsPanel.Focused = (p == PanelDetails)
	m.updateDetailsSelection()
}

func (m *AppModel) nextFocus() {
	next := (int(m.Focused) + 1) % 4
	m.setFocus(FocusedPanel(next))
}

func (m *AppModel) prevFocus() {
	prev := (int(m.Focused) + 3) % 4
	m.setFocus(FocusedPanel(prev))
}

func (m *AppModel) handleUp() {
	switch m.Focused {
	case PanelQueue:
		m.QueuePanel.MoveUp()
	case PanelCompleted:
		m.CompPanel.MoveUp()
	}
}

func (m *AppModel) handleDown() {
	switch m.Focused {
	case PanelQueue:
		m.QueuePanel.MoveDown()
	case PanelCompleted:
		m.CompPanel.MoveDown()
	}
}

func (m *AppModel) toggleFormat() {
	switch m.Cfg.Audio.Format {
	case "mp3":
		m.Cfg.Audio.Format = "flac"
	case "flac":
		m.Cfg.Audio.Format = "opus"
	case "opus":
		m.Cfg.Audio.Format = "copy"
	default:
		m.Cfg.Audio.Format = "mp3"
	}
	_ = m.Cfg.Save()
}

func (m *AppModel) updateDetailsSelection() {
	if m.Focused == PanelQueue {
		m.DetailsPanel.SetJob(m.QueuePanel.SelectedJob())
	} else if m.Focused == PanelCompleted {
		m.DetailsPanel.SetJob(m.CompPanel.SelectedJob())
	} else if m.DetailsPanel.Job == nil {
		if sel := m.QueuePanel.SelectedJob(); sel != nil {
			m.DetailsPanel.SetJob(sel)
		} else {
			m.DetailsPanel.SetJob(m.CompPanel.SelectedJob())
		}
	}
}

func (m *AppModel) updateSizes() {
	if m.Width <= 0 || m.Height <= 0 {
		return
	}

	contentHeight := m.Height - 1 // reserve 1 line for status bar
	if contentHeight < 10 {
		contentHeight = 10
	}

	leftColWidth := m.Width / 2
	if leftColWidth < 30 {
		leftColWidth = 30
	}
	rightColWidth := m.Width - leftColWidth

	configH := 8
	remH := contentHeight - configH
	queueH := remH / 2
	completedH := remH - queueH

	m.ConfigPanel.SetSize(leftColWidth, configH)
	m.QueuePanel.SetSize(leftColWidth, queueH)
	m.CompPanel.SetSize(leftColWidth, completedH)
	m.DetailsPanel.SetSize(rightColWidth, contentHeight)
}

// View renders application grid.
func (m AppModel) View() string {
	if m.Width == 0 || m.Height == 0 {
		return "Initializing..."
	}

	leftColumn := lipgloss.JoinVertical(
		lipgloss.Left,
		m.ConfigPanel.View(),
		m.QueuePanel.View(),
		m.CompPanel.View(),
	)

	rightColumn := m.DetailsPanel.View()

	grid := lipgloss.JoinHorizontal(
		lipgloss.Top,
		leftColumn,
		rightColumn,
	)

	statusBar := statusbar.View(m.Width, m.DependencyWarning)
	fullView := lipgloss.JoinVertical(lipgloss.Left, grid, statusBar)

	if m.ShowHelp {
		return helppopup.View(m.Width, m.Height)
	}
	if m.AddModal.Active {
		return m.AddModal.View(m.Width, m.Height)
	}
	if m.DirPicker.Active {
		return m.DirPicker.View(m.Width, m.Height)
	}
	if m.EditModal.Active {
		return m.EditModal.View(m.Width, m.Height)
	}

	return fullView
}
