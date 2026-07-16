package widget

import (
	"fmt"
	"strings"

	models "kanba/tui/models"

	"charm.land/lipgloss/v2"
)

type StatusBar struct {
	fileName   string
	fileIdx    int
	totalFiles int
	width      int
	theme      models.Theme
	copyMsg    string
}

func NewStatusBar(fileName string, fileIdx, totalFiles, width int, theme models.Theme, copyMsg string) *StatusBar {
	return &StatusBar{
		fileName:   fileName,
		fileIdx:    fileIdx,
		totalFiles: totalFiles,
		width:      width,
		theme:      theme,
		copyMsg:    copyMsg,
	}
}

func (s *StatusBar) Render() string {
	left := fmt.Sprintf(" ▸ %s  •  ↑↓ scroll  •  g/G top/bottom  •  ? help  •  q quit", s.fileName)
	right := ""
	if s.copyMsg != "" {
		right = s.copyMsg
	}

	text := left
	if right != "" {
		avail := s.width - lipgloss.Width(left) - 4
		if avail > len(right) {
			text = left + strings.Repeat(" ", avail-len(right)) + right
		}
	}

	style := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder(), false, false, true, false).
		BorderForeground(lipgloss.Color(s.theme.SidebarDir)).
		BorderBackground(lipgloss.Color(s.theme.SurfaceBg)).
		Foreground(lipgloss.Color(s.theme.StatusBarFg)).
		Background(lipgloss.Color(s.theme.SurfaceBg)).
		Padding(1, 1)
	return style.Width(s.width).Render(text)
}
