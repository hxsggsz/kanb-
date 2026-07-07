package tui

import (
	"fmt"
	"path/filepath"
	"strings"
	"unicode/utf8"

	"kanba/git"
)

type visualEntry struct {
	isDir   bool
	dir     string
	fileIdx int
	file    *git.SideBySideDiff
}

type Sidebar struct {
	files         []git.SideBySideDiff
	fileIdx       int
	width         int
	maxLines      int
	contentHeight int
}

func NewSidebar(files []git.SideBySideDiff, fileIdx int, width int, height int) *Sidebar {
	maxLines := max(height-statusBarHeight, 1)
	return &Sidebar{
		files:    files,
		fileIdx:  fileIdx,
		width:    width,
		maxLines: maxLines,
	}
}

func (s *Sidebar) Render() string {
	entries := buildVisualEntries(s.files)

	selPos := 0
	for i, e := range entries {
		if !e.isDir && e.fileIdx == s.fileIdx {
			selPos = i
			break
		}
	}

	start := 0
	if selPos >= s.maxLines {
		start = selPos - s.maxLines + 1
	}
	visible := entries[start:]
	if len(visible) > s.maxLines {
		visible = visible[:s.maxLines]
	}

	availableWidth := s.width - 3

	var sb strings.Builder
	lineCount := 0
	for _, e := range visible {
		if e.isDir {
			sb.WriteString(sidebarDirStyle.Render(truncate(e.dir, availableWidth)) + "\n")
		} else {
			statusColor := statusColorFor(e.file.Status)
			_, filename := filepath.Split(e.file.NewPath)
			if e.fileIdx == s.fileIdx {
				label := fmt.Sprintf("%s %s", statusColor.Render(e.file.Status), truncate(filename, availableWidth-4))
				sb.WriteString(sidebarFileSelected.Render("▸ "+label) + "\n")
			} else {
				label := fmt.Sprintf("%s %s", statusColor.Render(e.file.Status), truncate(filename, availableWidth-5))
				sb.WriteString(sidebarFile.Render("  "+label) + "\n")
			}
		}
		lineCount++
	}

	content := strings.TrimRight(sb.String(), "\n")
	s.contentHeight = lineCount
	return sidebarStyle.Width(s.width).Render(content)
}

func (s *Sidebar) ContentHeight() int { return s.contentHeight }

func CalculateSideWidth(totalWidth int) int {
	if totalWidth <= 0 {
		return sidebarDefaultWidth
	}
	return min(max(totalWidth/sidebarDenominator, sidebarMinWidth), sidebarMaxWidth)
}

func buildVisualEntries(files []git.SideBySideDiff) []visualEntry {
	var entries []visualEntry
	var lastDir string
	for i, f := range files {
		d := dirOf(f.NewPath)
		if d != lastDir {
			entries = append(entries, visualEntry{isDir: true, dir: d})
			lastDir = d
		}
		entries = append(entries, visualEntry{fileIdx: i, file: &files[i]})
	}
	return entries
}

func dirOf(path string) string {
	idx := strings.LastIndex(path, "/")
	if idx < 0 {
		return "./"
	}
	return path[:idx+1]
}

func truncate(s string, maxLen int) string {
	if maxLen < 0 {
		return ""
	}
	if utf8.RuneCountInString(s) <= maxLen {
		return s
	}
	if maxLen <= 3 {
		return string([]rune(s)[:maxLen])
	}
	return string([]rune(s)[:maxLen-3]) + "..."
}
