package tui

import (
	"fmt"
)

type StatusBar struct {
	fileIdx    int
	totalFiles int
	cursorLine int
	totalLines int
	width      int
}

func NewStatusBar(fileIdx, totalFiles, cursorLine, totalLines, width int) *StatusBar {
	return &StatusBar{
		fileIdx:    fileIdx,
		totalFiles: totalFiles,
		cursorLine: cursorLine,
		totalLines: totalLines,
		width:      width,
	}
}

func (s *StatusBar) Render() string {
	return statusBarStyle.Width(s.width).Render(
		fmt.Sprintf(" %d/%d  •  Ln %d/%d  •  ↑↓ cursor  •  n/p file  •  g/G top/bottom  •  ? help  •  q quit",
			s.fileIdx+1, s.totalFiles, s.cursorLine+1, s.totalLines))
}
