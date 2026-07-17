package app

import (
	"fmt"

	"kanba/tui/widget"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

type DefaultMode struct{}

func (m *DefaultMode) Type() ModeType { return ModeDefault }

func (m *DefaultMode) Render(model *Model) string {
	if len(model.flatLines) == 0 {
		return ""
	}

	theme := model.CurrentTheme()
	sideWidth := widget.CalculateSideWidth(model.width)
	scroll := model.scroller.Scroll()
	if scroll >= len(model.flatLines) {
		scroll = max(0, len(model.flatLines)-1)
	}
	cursorFileIdx := model.flatLines[scroll].FileIdx

	sidebar := widget.NewSidebar(model.diffs, cursorFileIdx, sideWidth, model.height, theme, model.fileStats)
	sidebarStr := sidebar.Render()

	contentVis := model.VisibleLines()
	panelWidth := max(model.width-sideWidth-panelBorderWidth, panelMinWidth)
	content := model.renderContinuous(panelWidth, contentVis)

	f := model.diffs[cursorFileIdx]
	statusBar := widget.NewStatusBar(f.NewPath, cursorFileIdx, len(model.diffs), model.width, theme, model.statusRightMsg())

	body := lipgloss.JoinHorizontal(lipgloss.Top, sidebarStr, content)
	result := fmt.Sprintf("%s\n%s", statusBar.Render(), body)

	theme = model.CurrentTheme()
	result = lipgloss.
		NewStyle().
		Width(model.width).
		Height(model.height).
		Background(lipgloss.Color(theme.PanelBg)).
		Render(result)
	result = model.themeModal.Overlay(result, theme.SurfaceBg, theme.SidebarSelected, theme.ContextFg)

	if model.helpActive {
		result = model.helpOverlay(result, theme)
	}

	return result
}

func (m *DefaultMode) HandleInput(model *Model, msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	return model.handleDiffKeys(msg)
}
