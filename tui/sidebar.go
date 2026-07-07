package tui

import (
	"fmt"
	"path/filepath"
	"strconv"
	"strings"
	"unicode/utf8"

	models "kanba/tui/models"

	"charm.land/lipgloss/v2"
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
	theme         models.Theme
	stats         []fileStat
}

func NewSidebar(files []git.SideBySideDiff, fileIdx int, width int, height int, theme models.Theme, stats []fileStat) *Sidebar {
	maxLines := max(height-statusBarHeight, 1)
	return &Sidebar{
		files:    files,
		fileIdx:  fileIdx,
		width:    width,
		maxLines: maxLines,
		theme:    theme,
		stats:    stats,
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

	sidebarStyle := lipgloss.NewStyle().
		Border(lipgloss.NormalBorder()).
		BorderRight(true).
		BorderLeft(false).
		BorderTop(false).
		BorderBottom(false).
		BorderForeground(lipgloss.Color(s.theme.BorderColor)).
		Background(lipgloss.Color(s.theme.PanelBg)).
		Padding(0, 1)

	sidebarFile := lipgloss.NewStyle().PaddingLeft(1)
	sidebarFileSelected := lipgloss.NewStyle().
		PaddingLeft(0).
		Foreground(lipgloss.Color(s.theme.SidebarSelected)).
		Bold(true)
	sidebarDirStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(s.theme.SidebarDir)).
		PaddingLeft(1)

	var sb strings.Builder
	lineCount := 0
	for _, e := range visible {
		if e.isDir {
			sb.WriteString(sidebarDirStyle.Render(truncate(e.dir, availableWidth)) + "\n")
		} else {
			statusColor := statusColorFor(e.file.Status, s.theme)
			_, filename := filepath.Split(e.file.NewPath)

			addStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(s.theme.SidebarAdded))
			delStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(s.theme.SidebarDeleted))
			st := s.stats[e.fileIdx]
			var statsParts []string
			if st.added > 0 {
				statsParts = append(statsParts, addStyle.Render("+"+strconv.Itoa(st.added)))
			}
			if st.deleted > 0 {
				statsParts = append(statsParts, delStyle.Render("-"+strconv.Itoa(st.deleted)))
			}
			statsStr := ""
			if len(statsParts) > 0 {
				statsStr = " (" + strings.Join(statsParts, ", ") + ")"
			}

			if e.fileIdx == s.fileIdx {
				label := fmt.Sprintf("%s %s%s", statusColor.Render(e.file.Status), truncate(filename, availableWidth-4-lipgloss.Width(statsStr)), statsStr)
				sb.WriteString(sidebarFileSelected.Render("▸ "+label) + "\n")
			} else {
				label := fmt.Sprintf("%s %s%s", statusColor.Render(e.file.Status), truncate(filename, availableWidth-5-lipgloss.Width(statsStr)), statsStr)
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
