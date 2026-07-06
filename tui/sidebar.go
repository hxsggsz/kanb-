package tui

import (
	"fmt"
	"strings"

	"kanba/git"
)

type Sidebar struct {
	files         []git.SideBySideDiff
	fileIdx       int
	width         int
	maxFiles      int
	contentHeight int
}

func NewSidebar(files []git.SideBySideDiff, fileIdx int, width int, height int) *Sidebar {
	maxFiles := max(height-(statusBarHeight+borderHeight), 1)
	return &Sidebar{
		files:    files,
		fileIdx:  fileIdx,
		width:    width,
		maxFiles: maxFiles,
	}
}

func (s *Sidebar) Render() string {
	start := 0
	if s.fileIdx >= s.maxFiles {
		start = s.fileIdx - s.maxFiles + 1
	}

	var sb strings.Builder
	for i := start; i < len(s.files) && i-start < s.maxFiles; i++ {
		f := s.files[i]
		statusColor := statusColorFor(f.Status)
		label := fmt.Sprintf("%s %s", statusColor.Render(f.Status), f.NewPath)
		if i == s.fileIdx {
			sb.WriteString(sidebarFileSelected.Render("▸ "+label) + "\n")
		} else {
			sb.WriteString(sidebarFile.Render("  "+label) + "\n")
		}
	}

	content := strings.TrimRight(sb.String(), "\n")
	s.contentHeight = strings.Count(content, "\n") + 1
	return sidebarStyle.Width(s.width).Render(content)
}

func (s *Sidebar) ContentHeight() int { return s.contentHeight }

func CalculateSideWidth(totalWidth int) int {
	if totalWidth <= 0 {
		return sidebarDefaultWidth
	}
	return min(max(totalWidth/sidebarDenominator, sidebarMinWidth), sidebarMaxWidth)
}
