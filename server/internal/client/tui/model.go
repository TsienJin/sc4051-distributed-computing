package tui

import (
	"fmt"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"server/internal/client"
)

type TabIndex int

const (
	FacilityMake TabIndex = iota
	FacilityQuery
	FacilityMonitor
	FacilityDelete

	BookingMake
	BookingModify
	BookingDelete

	LogsServer
	LogsClient
)

func (t TabIndex) MajorLabel() string {
	var label string
	switch t {
	case FacilityMake, FacilityQuery, FacilityMonitor, FacilityDelete:
		label = "Facility"
	case BookingMake, BookingModify, BookingDelete:
		label = "Booking"
	case LogsServer, LogsClient:
		label = "Logs"
	}
	return label
}

func (t TabIndex) MinorLabel() string {
	switch t {
	case FacilityMake:
		return "Make"
	case FacilityQuery:
		return "Query"
	case FacilityMonitor:
		return "Monitor"
	case FacilityDelete:
		return "Delete"
	case BookingMake:
		return "Make"
	case BookingModify:
		return "Modify"
	case BookingDelete:
		return "Delete"
	case LogsServer:
		return "Server"
	case LogsClient:
		return "Client"
	default:
		return "Unknown" // Failsafe
	}
}

func (t TabIndex) GetNeighbourMinor() []TabIndex {
	var res []TabIndex

	switch t {
	case FacilityMake, FacilityQuery, FacilityMonitor, FacilityDelete:
		res = append(res, FacilityMake, FacilityQuery, FacilityMonitor, FacilityDelete)
	case BookingMake, BookingModify, BookingDelete:
		res = append(res, BookingMake, BookingModify, BookingDelete)
	case LogsServer, LogsClient:
		res = append(res, LogsServer, LogsClient)
	}

	return res
}

func (t TabIndex) MajorLabelActive(currentTab TabIndex) bool {
	var active bool
	switch t {
	case FacilityMake, FacilityQuery, FacilityMonitor, FacilityDelete:
		active = currentTab <= FacilityDelete
	case BookingMake, BookingModify, BookingDelete:
		active = currentTab <= BookingDelete && FacilityDelete < currentTab
	case LogsServer, LogsClient:
		active = currentTab <= LogsClient && BookingDelete < currentTab
	}

	return active
}

func (t TabIndex) MajorIncr() TabIndex {
	switch t {
	case FacilityMake, FacilityQuery, FacilityMonitor, FacilityDelete:
		return BookingMake
	case BookingMake, BookingModify, BookingDelete:
		return LogsServer
	case LogsServer, LogsClient:
		return FacilityMake
	}
	return FacilityMake // Non-reachable code
}

func (t TabIndex) MajorDecr() TabIndex {
	switch t {
	case FacilityMake, FacilityQuery, FacilityMonitor, FacilityDelete:
		return LogsServer
	case BookingMake, BookingModify, BookingDelete:
		return FacilityMake
	case LogsServer, LogsClient:
		return BookingMake
	}
	return FacilityMake // Non-reachable code
}

func (t TabIndex) MinorDecr() TabIndex {
	switch t {
	case FacilityMake:
		return FacilityQuery
	case FacilityQuery:
		return FacilityMonitor
	case FacilityMonitor:
		return FacilityDelete
	case FacilityDelete:
		return FacilityMake // Wraps back to start of Facility group
	case BookingMake:
		return BookingModify
	case BookingModify:
		return BookingDelete
	case BookingDelete:
		return BookingMake // Wraps back to start of Booking group
	case LogsServer:
		return LogsClient
	case LogsClient:
		return LogsServer
	}
	return FacilityMake // Non-reachable, just a failsafe
}

func (t TabIndex) MinorIncr() TabIndex {
	switch t {
	case FacilityMake:
		return FacilityDelete // Wraps back to last in Facility group
	case FacilityQuery:
		return FacilityMake
	case FacilityMonitor:
		return FacilityQuery
	case FacilityDelete:
		return FacilityMonitor
	case BookingMake:
		return BookingDelete // Wraps back to last in Booking group
	case BookingModify:
		return BookingMake
	case BookingDelete:
		return BookingModify
	case LogsServer:
		return LogsClient
	case LogsClient:
		return LogsServer
	}
	return FacilityMake // Non-reachable, just a failsafe
}

type ViewMode int

const (
	ViewModeNav = iota
	ViewModeEdit
)

type MainModel struct {
	client *client.Client
	width  int
	height int

	tabIndex TabIndex
	viewMode ViewMode

	loading bool
}

func newModel(client *client.Client) MainModel {
	return MainModel{
		client:   client,
		width:    0,
		height:   0,
		viewMode: ViewModeNav,
		tabIndex: FacilityMake,
		loading:  false,
	}
}

func (m MainModel) Init() tea.Cmd {
	return nil
}

func (m MainModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		// Capture fullscreen width & height
		m.width = msg.Width
		m.height = msg.Height
		return m, nil
	case tea.KeyMsg:
		switch m.viewMode {
		case ViewModeNav:
			switch msg.String() {
			case "q": // Exit
				return m, tea.Quit
			case "left":
				m.tabIndex = m.tabIndex.MajorDecr()
				return m, nil
			case "right":
				m.tabIndex = m.tabIndex.MajorIncr()
				return m, nil
			case "up":
				m.tabIndex = m.tabIndex.MinorIncr()
				return m, nil
			case "down":
				m.tabIndex = m.tabIndex.MinorDecr()
				return m, nil
			}
		default:
			return m, cmd
		}
	}

	return m, cmd
}

func (m MainModel) View() string {
	// Create a styled area that fills the screen

	topBar := m.renderTopBar()
	sideBar := m.renderSideBar()
	mainElement := lipgloss.NewStyle().
		Width(m.width-lipgloss.Width(sideBar)).
		Height(m.height-lipgloss.Height(topBar)).
		Align(lipgloss.Center, lipgloss.Center).
		Border(lipgloss.HiddenBorder()).
		Padding(1)

	layout := lipgloss.JoinVertical(
		lipgloss.Top,
		topBar,
		lipgloss.JoinHorizontal(
			lipgloss.Top,
			sideBar,
			mainElement.Render(fmt.Sprintf(
				"Hello, World!\n(%d x %d)",
				m.width-lipgloss.Width(sideBar),
				m.height-lipgloss.Height(topBar),
			)),
		),
	)

	return layout
}
