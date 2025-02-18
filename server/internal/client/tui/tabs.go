package tui

import (
	"github.com/charmbracelet/lipgloss"
)

func (m MainModel) renderTopBarElement(item TabIndex) string {

	activeStyle := lipgloss.
		NewStyle().
		Bold(true).
		MarginTop(2).
		Padding(0).
		Width(30).
		Align(lipgloss.Center).
		Border(lipgloss.RoundedBorder())

	passiveStyle := lipgloss.
		NewStyle().
		Bold(false).
		MarginTop(2).
		Foreground(lipgloss.Color("248")).
		Width(30).
		Padding(0).
		Align(lipgloss.Center).
		BorderForeground(lipgloss.Color("240")).
		Border(lipgloss.RoundedBorder()).
		Faint(true)

	label := item.MajorLabel()
	active := item.MajorLabelActive(m.tabIndex)

	if active {
		return activeStyle.Render(label)
	}

	return passiveStyle.Render(label)
}

func (m MainModel) renderTopBarStatus() string {
	return ""
}

func (m MainModel) renderTopBar() string {

	facilityTab := m.renderTopBarElement(FacilityMake)
	bookingTab := m.renderTopBarElement(BookingMake)
	logTab := m.renderTopBarElement(LogsServer)

	return lipgloss.
		NewStyle().
		Width(m.width).
		Render(
			lipgloss.JoinHorizontal(
				lipgloss.Left,
				facilityTab,
				bookingTab,
				logTab,
			),
		)
}

func (m MainModel) renderSideBarElement(item TabIndex) string {

	passiveStyle := lipgloss.
		NewStyle().
		Width(20).
		Padding(1, 0).
		Margin(0, 1).
		Faint(true).
		Width(20).
		Align(lipgloss.Left)

	activeStyle := lipgloss.NewStyle().
		Faint(false).
		Padding(1, 0).
		Margin(0, 1).
		Foreground(lipgloss.Color("250")).
		Width(20).
		Align(lipgloss.Left).
		Bold(true)

	if m.tabIndex == item {
		return activeStyle.Render(item.MinorLabel())
	}

	return passiveStyle.Render(item.MinorLabel())
}

func (m MainModel) renderSideBar() string {

	// Get all neighbouring tabs
	neighbours := m.tabIndex.GetNeighbourMinor()

	// Render each element
	res := []string{}
	for _, n := range neighbours {
		res = append(res, m.renderSideBarElement(n))
	}

	return lipgloss.JoinVertical(lipgloss.Top, res...)
}
