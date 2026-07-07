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

	case diffMsg:
		m.loading = false
		if msg.err != nil {
			m.err = msg.err
		} else {
			m.diffs = msg.diffs
			m.flatLines = buildFlatLines(m.diffs)
			m.fileStats = computeFileStats(m.diffs)
		}
		return m, nil

	case tea.KeyPressMsg:
		return m.handleKeyPress(msg)
	}

	return m, nil
}

func (m *model) handleKeyPress(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	if m.err != nil {
		return m, tea.Quit
	}

	if m.themeModal.Active {
		switch msg.String() {
		case keyQuit, keyQuitAlt:
			m.themeModal.Close()
			return m, nil
		}
		return m.handleThemeModalKeys(msg)
	}

	switch msg.String() {
	case keyQuit, keyQuitAlt:
		return m, tea.Quit
	}

	switch m.screen {
	case screenDiff:
		return m.handleDiffKeys(msg)
	case screenHelp:
		if msg.String() == keyHelp || msg.String() == keyBack {
			m.screen = screenDiff
		}
		return m, nil
	}

	return m, nil
}

func (m *model) handleDiffKeys(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case keyUp, keyUpAlt:
		m.scroller.MoveUp()

	case keyDown, keyDownAlt:
		m.scroller.MoveDown(m.totalLines())

	case keyTop:
		m.scroller.GoToTop()

	case keyBottom:
		m.scroller.GoToBottom(m.totalLines())

	case keyLeft, keyLeftAlt:
		m.scroller.ScrollLeft()

	case keyRight, keyRightAlt:
		m.scroller.ScrollRight()

	case keyLeftWord:
		m.scroller.ScrollLeftFast()

	case keyRightWord:
		m.scroller.ScrollRightFast()

	case keyHome:
		m.scroller.ScrollHome()

	case keyEnd:
		sideWidth := CalculateSideWidth(m.width)
		panelWidth := max(m.width-sideWidth-panelBorderWidth, panelMinWidth)
		colWidth := (panelWidth - 3) / 2
		prefixWidth := lineNumColWidth + 3
		maxContent := 0
		for _, f := range m.diffs {
			w := maxFileContentWidth(f)
			if w > maxContent {
				maxContent = w
			}
		}
		m.scroller.ScrollEnd(max(0, maxContent-(colWidth-prefixWidth)))

	case keyHelp:
		m.screen = screenHelp

	case keyTheme:
		m.themeModal.Active = true
		m.themeModal.SyncCursor(m.themeModal.Selected)
	}

	return m, nil
}

func (m *model) handleThemeModalKeys(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case keyUp, keyUpAlt:
		m.themeModal.MoveUp()
	case keyDown, keyDownAlt:
		m.themeModal.MoveDown()
	case "enter":
		m.themeModal.Select()
	case keyTheme, keyBack, keyQuitAlt:
		m.themeModal.Close()
	}
	return m, nil
}
