package tui

import (
	tea "charm.land/bubbletea/v2"
)

func (m *model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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

func (m *model) handleKeyPress(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case keyQuit, keyQuitAlt:
		return m, tea.Quit
	}

	switch m.screen {
	case screenDiff:
		return m.handleDiffKeys(msg)
	case screenHelp:
		return m.handleHelpKeys(msg)
	}

	return m, nil
}

func (m *model) handleDiffKeys(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	return m, nil
}

func (m *model) handleHelpKeys(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	if msg.String() == keyBack || msg.String() == keyHelp {
		m.screen = screenDiff
	}
	return m, nil
}
