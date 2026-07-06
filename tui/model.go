package tui

import "charm.land/bubbletea/v2"

type screen int

const (
	screenHome screen = iota
	screenDetail
	screenHelp
)

type model struct {
	currentScreen screen
	cursor        int
	items         []item
	loading       bool
	err           error
	width         int
	height        int
}

type item struct {
	title       string
	description string
}

func initialModel() model {
	return model{
		currentScreen: screenHome,
		items: []item{
			{title: "Welcome to Kanba", description: "A Bubble Tea v2 starter"},
			{title: "Arrow keys", description: "Navigate the list"},
			{title: "Enter", description: "View details"},
			{title: "?", description: "Show help"},
		},
	}
}

func New() tea.Model {
	return initialModel()
}

func (m model) Init() tea.Cmd {
	return nil
}
