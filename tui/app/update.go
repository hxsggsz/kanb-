package app

import (
	"kanba/tui/diff"
	"kanba/tui/widget"
	"kanba/tui/setting"

	tea "charm.land/bubbletea/v2"
)

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case setting.DiffMsg:
		m.loading = false
		if msg.Err != nil {
			m.err = msg.Err
		} else {
			m.diffs = msg.Diffs
			m.flatLines = diff.BuildFlatLines(m.diffs)
			m.fileStats = diff.ComputeFileStats(m.diffs)
		}
		return m, nil

	case tea.KeyPressMsg:
		return m.handleKeyPress(msg)
	}

	return m, nil
}

func (m *Model) handleKeyPress(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	if m.err != nil {
		return m, tea.Quit
	}

	if m.themeModal.Active {
		switch msg.String() {
		case setting.KeyQuit, setting.KeyQuitAlt:
			m.themeModal.Close()
			return m, nil
		}
		return m.handleThemeModalKeys(msg)
	}

	switch msg.String() {
	case setting.KeyQuit, setting.KeyQuitAlt:
		return m, tea.Quit
	}

	switch m.screen {
	case screenDiff:
		return m.handleDiffKeys(msg)
	case screenHelp:
		if msg.String() == setting.KeyHelp || msg.String() == setting.KeyBack {
			m.screen = screenDiff
		}
		return m, nil
	}

	return m, nil
}

func (m *Model) handleDiffKeys(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case setting.KeyUp, setting.KeyUpAlt:
		m.scroller.MoveUp()
		for m.scroller.CursorLine() > 0 && m.flatLines[m.scroller.CursorLine()].IsHeader {
			m.scroller.MoveUp()
		}

	case setting.KeyDown, setting.KeyDownAlt:
		m.scroller.MoveDown(m.TotalLines())
		for m.scroller.CursorLine() < m.TotalLines()-1 && m.flatLines[m.scroller.CursorLine()].IsHeader {
			m.scroller.MoveDown(m.TotalLines())
		}

	case setting.KeyTop:
		m.scroller.GoToTop()
		for m.scroller.CursorLine() < m.TotalLines()-1 && m.flatLines[m.scroller.CursorLine()].IsHeader {
			m.scroller.MoveDown(m.TotalLines())
		}

	case setting.KeyBottom:
		m.scroller.GoToBottom(m.TotalLines())
		for m.scroller.CursorLine() > 0 && m.flatLines[m.scroller.CursorLine()].IsHeader {
			m.scroller.MoveUp()
		}

	case setting.KeyLeft, setting.KeyLeftAlt:
		m.scroller.ScrollLeft()

	case setting.KeyRight, setting.KeyRightAlt:
		m.scroller.ScrollRight()

	case setting.KeyLeftWord:
		m.scroller.ScrollLeftFast()

	case setting.KeyRightWord:
		m.scroller.ScrollRightFast()

	case setting.KeyHome:
		m.scroller.ScrollHome()

	case setting.KeyEnd:
		sideWidth := widget.CalculateSideWidth(m.width)
		panelWidth := max(m.width-sideWidth-panelBorderWidth, panelMinWidth)
		colWidth := (panelWidth - 3) / 2
		prefixWidth := diff.LineNumColWidth + 3
		maxContent := 0
		for _, f := range m.diffs {
			w := maxFileContentWidth(f)
			if w > maxContent {
				maxContent = w
			}
		}
		m.scroller.ScrollEnd(max(0, maxContent-(colWidth-prefixWidth)))

	case setting.KeyHelp:
		m.screen = screenHelp

	case setting.KeyTheme:
		m.themeModal.Active = true
		m.themeModal.SyncCursor(m.themeModal.Selected)
	}

	return m, nil
}

func (m *Model) handleThemeModalKeys(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case setting.KeyUp, setting.KeyUpAlt:
		m.themeModal.MoveUp()
	case setting.KeyDown, setting.KeyDownAlt:
		m.themeModal.MoveDown()
	case "enter":
		m.themeModal.Select()
	case setting.KeyTheme, setting.KeyBack, setting.KeyQuitAlt:
		m.themeModal.Close()
	}
	return m, nil
}
