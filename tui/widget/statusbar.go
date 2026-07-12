package widget

import (
	"fmt"

	models "kanba/tui/models"

	"charm.land/lipgloss/v2"
)

type StatusBar struct {
	fileName   string
	fileIdx    int
	totalFiles int
	width      int
	theme      models.Theme
}

func NewStatusBar(fileName string, fileIdx, totalFiles, width int, theme models.Theme) *StatusBar {
	return &StatusBar{
		fileName:   fileName,
		fileIdx:    fileIdx,
		totalFiles: totalFiles,
		width:      width,
		theme:      theme,
	}
}

func (s *StatusBar) Render() string {
	style := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder(), false, false, true, false).
		BorderForeground(lipgloss.Color(s.theme.SidebarDir)).
		BorderBackground(lipgloss.Color(s.theme.SurfaceBg)).
		Foreground(lipgloss.Color(s.theme.StatusBarFg)).
		Background(lipgloss.Color(s.theme.SurfaceBg)).
		Padding(1, 1)
	return style.Width(s.width).Render(
		fmt.Sprintf(" ▸ %s  •  ↑↓ scroll  •  g/G top/bottom  •  ? help  •  q quit",
			s.fileName))
}
