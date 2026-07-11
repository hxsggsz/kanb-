package app

import (
	"log/slog"

	"kanba/tui/diff"
	"kanba/tui/selection"
	"kanba/tui/setting"
	"kanba/tui/widget"

	"charm.land/lipgloss/v2"
	tea "charm.land/bubbletea/v2"
)

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.activeMode = m.modeFactory.FromWidth(msg.Width)
		return m, nil

	case setting.DiffMsg:
		m.loading = false
		if msg.Err != nil {
			m.err = msg.Err
		} else {
			m.diffs = msg.Diffs
			m.flatLines = diff.BuildFlatLines(m.diffs)
			m.fileStats = diff.ComputeFileStats(m.diffs)
			m.setupSelectionProvider()
		}
		return m, nil

	case tea.KeyPressMsg:
		return m.handleKeyPress(msg)

	case tea.MouseClickMsg:
		return m.handleMouseClick(msg), nil
	case tea.MouseWheelMsg:
		return m.handleMouseWheel(msg), nil

	case tea.MouseMotionMsg:
		if m.selection != nil {
			panel, line, col := m.mapMouseToContent(msg.X, msg.Y)
			if panel >= 0 {
				m.selection.HandleDrag(panel, line, col)
			}
		}

	case tea.MouseReleaseMsg:
		if m.selection != nil {
			return m, m.selection.HandleRelease()
		}

	case selection.CopyMsg:
		if m.selectedText == "" {
			return m, nil
		}
		if err := selection.CopyToClipboard(m.selectedText); err != nil {
			slog.Warn("failed to copy to clipboard", "error", err)
		}
		m.selectedText = ""
		return m, nil
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

	if m.helpActive {
		switch msg.String() {
		case setting.KeyHelp, setting.KeyBack:
			m.helpActive = false
		}
		return m, nil
	}

	switch msg.String() {
	case setting.KeyQuit, setting.KeyQuitAlt:
		return m, tea.Quit
	}

	if m.activeMode != nil {
		return m.activeMode.HandleInput(m, msg)
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
		m.helpActive = true

	case setting.KeyTheme:
		m.themeModal.Active = true
		m.themeModal.SyncCursor(m.themeModal.Selected)
	}

	return m, nil
}

func (m *Model) handleMouseWheel(msg tea.MouseWheelMsg) *Model {
	if len(m.flatLines) == 0 {
		return m
	}
	if m.themeModal.Active {
		switch msg.Button {
		case tea.MouseWheelUp:
			m.themeModal.MoveUp()
		case tea.MouseWheelDown:
			m.themeModal.MoveDown()
		}
		return m
	}

	switch msg.Button {
	case tea.MouseWheelUp:
		if msg.Mod.Contains(tea.ModShift) {
			m.scroller.ScrollLeft()
		} else {
			m.scroller.ScrollViewBy(-mouseScrollSpeed, m.TotalLines())
		}
	case tea.MouseWheelDown:
		if msg.Mod.Contains(tea.ModShift) {
			m.scroller.ScrollRight()
		} else {
			m.scroller.ScrollViewBy(mouseScrollSpeed, m.TotalLines())
		}
	case tea.MouseWheelLeft:
		m.scroller.ScrollLeft()
	case tea.MouseWheelRight:
		m.scroller.ScrollRight()
	}
	return m
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

func (m *Model) handleMouseClick(msg tea.MouseClickMsg) *Model {
	if len(m.flatLines) == 0 {
		return m
	}
	x, y := msg.X, msg.Y

	if m.themeModal.Active {
		theme := m.CurrentTheme()
		sideWidth := widget.CalculateSideWidth(m.width)
		panelWidth := max(m.width-sideWidth-panelBorderWidth, panelMinWidth)

		fg := m.themeModal.Render(theme.PanelBg, theme.SidebarSelected, theme.ContextFg)
		fgWidth, fgHeight := lipgloss.Size(fg)
		modalX := sideWidth + max(0, (panelWidth-fgWidth)/2)
		modalY := max(0, (m.height-fgHeight)/2)

		if x >= modalX && x < modalX+fgWidth && y >= modalY && y < modalY+fgHeight {
			relY := y - modalY
			if relY >= 4 && relY < 4+len(m.themeModal.Items) {
				idx := relY - 4
				m.themeModal.Cursor = idx
				m.themeModal.Select()
				m.themeModal.SyncCursor(m.themeModal.Selected)
			}
		} else {
			m.themeModal.Close()
		}
		return m
	}

	if m.helpActive {
		theme := m.CurrentTheme()
		sideWidth := widget.CalculateSideWidth(m.width)
		panelWidth := max(m.width-sideWidth-panelBorderWidth, panelMinWidth)

		fg := m.helpContent(theme)
		fgWidth, fgHeight := lipgloss.Size(fg)
		modalX := sideWidth + max(0, (panelWidth-fgWidth)/2)
		modalY := max(0, m.height/2-fgHeight/2)

		if !(x >= modalX && x < modalX+fgWidth && y >= modalY && y < modalY+fgHeight) {
			m.helpActive = false
		}
		return m
	}

	if y < statusBarHeight {
		return m
	}

	contentY := y - statusBarHeight
	sideWidth := widget.CalculateSideWidth(m.width)

	if x < sideWidth {
		fileIdx, ok := widget.LookupSidebarEntry(m.diffs, m.flatLines[m.scroller.CursorLine()].FileIdx, m.height, contentY)
		if !ok {
			return m
		}
		for i, fl := range m.flatLines {
			if fl.FileIdx == fileIdx && !fl.IsHeader {
				m.scroller.GoToTop()
				for m.scroller.CursorLine() < i {
					m.scroller.MoveDown(len(m.flatLines))
				}
				return m
			}
		}
		return m
	}

	panel, line, col := m.mapMouseToContent(msg.X, msg.Y)
	isClickInContent := panel >= 0
	hasSelection := m.selection != nil
	if isClickInContent && hasSelection {
		m.selection.HandleClick(panel, line, col)
	}

	start := m.scroller.Scroll()
	vis := m.VisibleLines()
	total := m.TotalLines()
	end := min(start+vis, total)
	visualRow := 0
	targetLine := end - 1
	for gi := start; gi < end; gi++ {
		fl := m.flatLines[gi]
		h := 1
		if fl.IsHeader {
			h = 3
			if fl.FileIdx > 0 {
				h = 4
			}
		}
		if visualRow+h > contentY {
			targetLine = gi
			break
		}
		visualRow += h
	}
	m.scroller.GoToTop()
	for m.scroller.CursorLine() < targetLine {
		m.scroller.MoveDown(len(m.flatLines))
	}
	for m.scroller.CursorLine() < len(m.flatLines)-1 && m.flatLines[m.scroller.CursorLine()].IsHeader {
		m.scroller.MoveDown(len(m.flatLines))
	}
	return m
}

// mapMouseToContent maps mouse coordinates to panel, line, and column.
// Returns panel (-1 if outside content area), flat line index, and content column.
func (m *Model) mapMouseToContent(x, y int) (selection.PanelSide, int, int) {
	if y < statusBarHeight {
		return -1, 0, 0
	}

	contentY := y - statusBarHeight
	sideWidth := widget.CalculateSideWidth(m.width)

	isInsideSidebar := x < sideWidth
	if isInsideSidebar {
		return -1, 0, 0
	}

	// Determine panel
	panelWidth := max(m.width-sideWidth-panelBorderWidth, panelMinWidth)
	colWidth := panelWidth / 2
	contentX := x - sideWidth
	isLeftPanel := contentX < colWidth

	var panel selection.PanelSide
	var panelLeft int
	if isLeftPanel {
		panel = selection.PanelLeft
		panelLeft = 0
	} else {
		panel = selection.PanelRight
		panelLeft = colWidth
	}

	// Map y to flat line index
	start := m.scroller.Scroll()
	vis := m.VisibleLines()
	total := m.TotalLines()
	end := min(start+vis, total)
	visualRow := 0
	targetLine := end - 1
	for gi := start; gi < end; gi++ {
		fl := m.flatLines[gi]
		h := m.lineVisualHeight(fl)
		clickFallsInLine := visualRow+h > contentY
		if clickFallsInLine {
			targetLine = gi
			break
		}
		visualRow += h
	}

	// Skip headers
	for targetLine < len(m.flatLines) && m.flatLines[targetLine].IsHeader {
		targetLine++
	}
	isOutOfBounds := targetLine >= len(m.flatLines)
	if isOutOfBounds {
		return -1, 0, 0
	}

	// Map x to content column
	prefixWidth := diff.LineNumColWidth + 3
	contentCol := contentX - panelLeft - prefixWidth + m.scroller.HScroll()
	if contentCol < 0 {
		contentCol = 0
	}

	return panel, targetLine, contentCol
}

// lineVisualHeight returns the visual height of a flat line in rows.
func (m *Model) lineVisualHeight(fl diff.FlatLine) int {
	if !fl.IsHeader {
		return 1
	}
	if fl.FileIdx > 0 {
		return 4
	}
	return 3
}
