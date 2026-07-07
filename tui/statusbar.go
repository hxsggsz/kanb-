package tui

import (
	"fmt"

	models "kanba/tui/models"

	"charm.land/lipgloss/v2"
)

type StatusBar struct {
	fileName   string
	fileIdx    int
	totalFiles int
	cursorLine int
	totalLines int
	width      int
	theme      models.Theme
}

func NewStatusBar(fileName string, fileIdx, totalFiles, cursorLine, totalLines, width int, theme models.Theme) *StatusBar {
	return &StatusBar{
		fileName:   fileName,
		fileIdx:    fileIdx,
		totalFiles: totalFiles,
		cursorLine: cursorLine,
		totalLines: totalLines,
		width:      width,
		theme:      theme,
	}
}

func (s *StatusBar) Render() string {
	style := lipgloss.NewStyle().
		Foreground(lipgloss.Color(s.theme.StatusBarFg)).
		Background(lipgloss.Color(s.theme.StatusBarBg)).
		Padding(1, 1)
	return style.Width(s.width).Render(
		fmt.Sprintf(" ▸ %s  •  Ln %d/%d  •  ↑↓ cursor  •  g/G top/bottom  •  ? help  •  q quit",
			s.fileName, s.cursorLine+1, s.totalLines))
}
