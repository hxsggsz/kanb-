package tui

import (
	tea "charm.land/bubbletea/v2"
)

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case tea.KeyPressMsg:
		return m.handleKeyPress(msg)

	case statusMsg:
		m.loading = false
		return m, nil

	case errMsg:
		m.err = msg
		m.loading = false
		return m, nil

	case tickMsg:
		return m, nil
	}

	return m, nil
}

func (m model) handleKeyPress(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case keyQuit, keyQuitAlt:
		return m, tea.Quit
	}

	switch m.currentScreen {
	case screenHome:
		return m.handleHomeKeys(msg)
	case screenDetail:
		return m.handleDetailKeys(msg)
	case screenHelp:
		return m.handleHelpKeys(msg)
	}

	return m, nil
}

func (m model) handleHomeKeys(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case keyUp, keyUpAlt:
		if m.cursor > 0 {
			m.cursor--
		}
	case keyDown, keyDownAlt:
		if m.cursor < len(m.items)-1 {
			m.cursor++
		}
	case keyEnter:
		m.currentScreen = screenDetail
	case keyHelp:
		m.currentScreen = screenHelp
	case keyRefresh:
		m.loading = true
		return m, checkServer("https://example.com")
	}

	return m, nil
}

func (m model) handleDetailKeys(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	if msg.String() == keyBack {
		m.currentScreen = screenHome
	}
	return m, nil
}

func (m model) handleHelpKeys(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	if msg.String() == keyBack || msg.String() == keyHelp {
		m.currentScreen = screenHome
	}
	return m, nil
}
