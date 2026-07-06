package tui

import (
	"fmt"
	"strconv"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"kanba/git"
)

func (m *model) View() tea.View {
	v := tea.NewView("")
	v.AltScreen = true
	v.WindowTitle = "kanba"

	if m.loading {
		v.SetContent(m.loadingView())
		return v
	}
	if m.err != nil {
		v.SetContent(m.errorView())
		return v }
	if len(m.diffs) == 0 {
		v.SetContent(m.emptyView())
		return v
	}

	switch m.screen {
	case screenDiff:
		v.SetContent(m.diffView())
	case screenHelp:
		v.SetContent(m.helpView())
	}

	return v
}

func (m *model) loadingView() string {
	return loadingStyle.Render(" Loading diffs...")
}

func (m *model) errorView() string {
	return errorStyle.Render(fmt.Sprintf("Error: %v\n\nPress any key to exit.", m.err))
}

func (m *model) emptyView() string {
	return loadingStyle.Render(" No changes to show.")
}

func (m *model) diffView() string {
	sideWidth := sidebarDefaultWidth
	if m.width > 0 {
		sideWidth = min(max(m.width/sidebarDenominator, sidebarMinWidth), sidebarMaxWidth)
	}

	maxFiles := max(m.height - (statusBarHeight + borderHeight), 1)

	start := 0
	if m.fileIdx >= maxFiles {
		start = m.fileIdx - maxFiles + 1
	}

	var sb strings.Builder
	for i := start; i < len(m.diffs) && i-start < maxFiles; i++ {
		f := m.diffs[i]
		statusColor := statusColorFor(f.Status)
		label := fmt.Sprintf("%s %s", statusColor.Render(f.Status), f.NewPath)
		if i == m.fileIdx {
			sb.WriteString(sidebarFileSelected.Render("▸ "+label) + "\n")
		} else {
			sb.WriteString(sidebarFile.Render("  "+label) + "\n")
		}
	}
	sidebarContent := strings.TrimRight(sb.String(), "\n")
	sidebar := sidebarStyle.Width(sideWidth).Render(sidebarContent)

	sidebarHeight := strings.Count(sidebarContent, "\n") + 1
	contentVis := max(m.height - sidebarHeight - 2, 1)

	panelWidth := max(m.width - sideWidth - panelBorderWidth, panelMinWidth)

	file := m.diffs[m.fileIdx]
	content := m.renderFile(file, panelWidth, contentVis)

	total := m.totalLines()
	cursorLine := m.cursorLine + 1
	statusBar := statusBarStyle.Width(m.width).Render(
		fmt.Sprintf(" %d/%d  •  Ln %d/%d  •  ↑↓ cursor  •  n/p file  •  g/G top/bottom  •  ? help  •  q quit",
			m.fileIdx+1, len(m.diffs), cursorLine, total))

	return fmt.Sprintf("%s%s\n%s", sidebar, content, statusBar)
}

func (m *model) renderFile(f git.FileDiff, width int, vis int) string {
	total := m.totalLines()
	if total == 0 {
		return ""
	}

	if m.cursorLine >= total {
		m.cursorLine = total - 1
	}

	const scrollMargin = 8

	if vis <= 0 {
		m.scroll = 0
		return ""
	}

	if vis >= total {
		m.scroll = 0
	} else {
		sm := scrollMargin
		if sm > vis/2 {
			sm = max(1, vis/2)
		}
		maxScroll := total - vis
		if m.cursorLine < m.scroll+sm {
			m.scroll = max(0, m.cursorLine-sm)
		}
		if m.cursorLine >= m.scroll+vis-sm {
			m.scroll = min(m.cursorLine-vis+sm+1, maxScroll)
		}
		if m.scroll > maxScroll {
			m.scroll = maxScroll
		}
	}

	var lines []string
	lineIdx := 0
	for _, h := range f.Hunks {
		cursor := lineIdx == m.cursorLine
		line := hunkHeaderStyle.Render(h.Header)
		if cursor {
			line = lineCursorStyle.Width(width).Render(line)
		}
		lines = append(lines, line)
		lineIdx++

		for _, ln := range h.Lines {
			cursor := lineIdx == m.cursorLine
			lines = append(lines, formatLine(ln, width, cursor))
			lineIdx++
		}
	}

	start := m.scroll
	end := start + vis
	if start >= len(lines) {
		start = 0
	}
	if end > len(lines) {
		end = len(lines)
	}

	return strings.Join(lines[start:end], "\n")
}

func formatLine(ln git.Line, width int, cursor bool) string {
	oldStr := ""
	newStr := ""
	if ln.OldLineNum > 0 {
		oldStr = strconv.Itoa(ln.OldLineNum)
	}
	if ln.NewLineNum > 0 {
		newStr = strconv.Itoa(ln.NewLineNum)
	}

	lineNumFmt := fmt.Sprintf("%%%ds %%%ds", lineNumColWidth, lineNumColWidth)
	lineNum := fmt.Sprintf(lineNumFmt, oldStr, newStr)

	prefix := " "
	switch ln.Type {
	case git.LineAdded:
		prefix = "+"
	case git.LineDeleted:
		prefix = "-"
	}

	if cursor {
		line := fmt.Sprintf("%s %s %s", lineNum, prefix, ln.Content)
		if len(line) > width {
			line = line[:width]
		}
		style := lipgloss.NewStyle().Background(lipgloss.Color("#444444"))
		switch ln.Type {
		case git.LineAdded:
			style = style.Foreground(lipgloss.Color("#00FF00"))
		case git.LineDeleted:
			style = style.Foreground(lipgloss.Color("#FF0000"))
		}
		return style.Render(line)
	}

	style := lineContextStyle
	switch ln.Type {
	case git.LineAdded:
		style = lineAddedStyle
	case git.LineDeleted:
		style = lineDeletedStyle
	}

	line := fmt.Sprintf("%s %s %s", lineNumStyle.Render(lineNum), prefix, ln.Content)
	if len(line) > width {
		line = line[:width]
	}

	return style.Render(line)
}

func statusColorFor(status string) lipgloss.Style {
	switch status {
	case "A":
		return sidebarStatusAdded
	case "D":
		return sidebarStatusDeleted
	default:
		return sidebarStatusModified
	}
}

func (m *model) helpView() string {
	var content strings.Builder; content.WriteString(" Keybindings\n\n")
	bindings := []struct{ key, desc string }{
		{"↑/k", "Cursor up"},
		{"↓/j", "Cursor down"},
		{"n", "Next file"},
		{"p", "Previous file"},
		{"g", "Go to top"},
		{"G", "Go to bottom"},
		{"?", "Toggle help"},
		{"q/ctrl+c", "Quit"},
	}
	for _, b := range bindings {
		content .WriteString(fmt.Sprintf("  %-12s %s\n", b.key, b.desc))
	}
	return helpStyle.Render(content.String())
}
