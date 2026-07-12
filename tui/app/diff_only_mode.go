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

	scroll := model.scroller.Scroll()
	if scroll >= len(model.flatLines) {
		scroll = max(0, len(model.flatLines)-1)
	}
	cursorFileIdx := model.flatLines[scroll].FileIdx
	f := model.diffs[cursorFileIdx]
	statusBar := widget.NewStatusBar(f.NewPath, cursorFileIdx, len(model.diffs), model.width, theme)

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
