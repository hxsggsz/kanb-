package app

import (
	"fmt"

	"kanba/tui/widget"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

type DiffOnlyMode struct{}

func (m *DiffOnlyMode) Type() ModeType { return ModeDiffOnly }

func (m *DiffOnlyMode) Render(model *Model) string {
	if len(model.flatLines) == 0 {
		return ""
	}

	theme := model.CurrentTheme()
	contentVis := model.VisibleLines()
	panelWidth := max(model.width-panelBorderWidth, panelMinWidth)
	content := model.renderContinuous(panelWidth, contentVis)

	cursorFileIdx := model.flatLines[model.scroller.CursorLine()].FileIdx
	f := model.diffs[cursorFileIdx]
	statusBar := widget.NewStatusBar(f.NewPath, cursorFileIdx, len(model.diffs), model.scroller.CursorLine(), len(model.flatLines), model.width, theme)

	result := fmt.Sprintf("%s\n%s", statusBar.Render(), content)
	result = lipgloss.NewStyle().Background(lipgloss.Color(theme.PanelBg)).Render(result)
	result = model.themeModal.Overlay(result, theme.PanelBg, theme.SidebarSelected, theme.ContextFg, 0, panelWidth)

	if model.helpActive {
		result = model.helpOverlay(result, theme, 0, panelWidth)
	}

	return result
}

func (m *DiffOnlyMode) HandleInput(model *Model, msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	return model.handleDiffKeys(msg)
}
