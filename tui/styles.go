package tui

import "charm.land/lipgloss/v2"

var (
	sidebarStyle = lipgloss.NewStyle().
			Border(lipgloss.NormalBorder()).
			BorderRight(true).
			Padding(0, 1).
			Width(30)

	sidebarFile = lipgloss.NewStyle().
			PaddingLeft(1)

	sidebarFileSelected = lipgloss.NewStyle().
				PaddingLeft(0).
				Foreground(lipgloss.Color("#7D56F4")).
				Bold(true)

	sidebarStatusAdded = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#00FF00"))

	sidebarStatusDeleted = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#FF0000"))

	sidebarStatusModified = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#FFA500"))

	lineAddedStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#00FF00")).
			Background(lipgloss.Color("#003300"))

	lineDeletedStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FF0000")).
			Background(lipgloss.Color("#330000"))

	lineContextStyle = lipgloss.NewStyle()

	lineNumStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#888888"))

	hunkHeaderStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#8888FF")).
			Bold(true)

	statusBarStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FAFAFA")).
			Background(lipgloss.Color("#333333")).
			Padding(0, 1)

	helpStyle = lipgloss.NewStyle().
			Padding(1, 2)

	errorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FF0000")).
			Bold(true).
			Padding(2, 4)

	loadingStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#7D56F4")).
			Padding(2, 4)

	appStyle = lipgloss.NewStyle().
			Padding(1, 2)

	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FAFAFA")).
			Background(lipgloss.Color("#7D56F4")).
			Padding(0, 1)

	itemStyle = lipgloss.NewStyle().
			PaddingLeft(2)

	selectedItemStyle = lipgloss.NewStyle().
				PaddingLeft(1).
				Foreground(lipgloss.Color("#7D56F4")).
				Bold(true)

	detailStyle = lipgloss.NewStyle().
			Border(lipgloss.NormalBorder()).
			Padding(1, 2)

	spinnerStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#7D56F4"))
)
