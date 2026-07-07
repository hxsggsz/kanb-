package tui

import (
	"fmt"
)

type StatusBar struct {
	fileName   string
	fileIdx    int
	totalFiles int
	cursorLine int
	totalLines int
	width      int
}

func NewStatusBar(fileName string, fileIdx, totalFiles, cursorLine, totalLines, width int) *StatusBar {
	return &StatusBar{
		fileName:   fileName,
		fileIdx:    fileIdx,
		totalFiles: totalFiles,
		cursorLine: cursorLine,
		totalLines: totalLines,
		width:      width,
	}
}

func (s *StatusBar) Render() string {
	return statusBarStyle.Width(s.width).Render(
		fmt.Sprintf(" ▸ %s  •  Ln %d/%d  •  ↑↓ cursor  •  g/G top/bottom  •  ? help  •  q quit",
			s.fileName, s.cursorLine+1, s.totalLines))
}
