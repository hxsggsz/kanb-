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

	body := lipgloss.JoinHorizontal(lipgloss.Top, sidebar, content)
	return fmt.Sprintf("%s\n%s", body, statusBar)
}

func (m *model) renderFile(f git.SideBySideDiff, width int, vis int) string {
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

	colWidth := (width - 3) / 2

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
			lines = append(lines, formatAlignedLine(ln, colWidth, cursor))
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

func formatAlignedLine(ln git.AlignedLine, colWidth int, cursor bool) string {
	oldNum := ""
	if ln.OldLineNum > 0 {
		oldNum = strconv.Itoa(ln.OldLineNum)
	}
	newNum := ""
	if ln.NewLineNum > 0 {
		newNum = strconv.Itoa(ln.NewLineNum)
	}

	leftPrefix := " "
	leftContent := ""
	if ln.Kind == git.KindContext || ln.Kind == git.KindDeleted || ln.Kind == git.KindModified {
		leftContent = ln.OldContent
	}
	if ln.Kind == git.KindDeleted || ln.Kind == git.KindModified {
		leftPrefix = "-"
	}

	rightPrefix := " "
	rightContent := ""
	if ln.Kind == git.KindContext || ln.Kind == git.KindAdded || ln.Kind == git.KindModified {
		rightContent = ln.NewContent
	}
	if ln.Kind == git.KindAdded || ln.Kind == git.KindModified {
		rightPrefix = "+"
	}

	leftLine := fmt.Sprintf("%*s %s %s", lineNumColWidth, oldNum, leftPrefix, leftContent)
	rightLine := fmt.Sprintf("%*s %s %s", lineNumColWidth, newNum, rightPrefix, rightContent)

	var leftStyle, rightStyle lipgloss.Style
	if cursor {
		base := lipgloss.NewStyle().Background(lipgloss.Color("#444444"))
		leftStyle = base
		rightStyle = base
		switch ln.Kind {
		case git.KindAdded:
			rightStyle = base.Foreground(lipgloss.Color("#00FF00"))
		case git.KindDeleted:
			leftStyle = base.Foreground(lipgloss.Color("#FF0000"))
		case git.KindModified:
			leftStyle = base.Foreground(lipgloss.Color("#FF0000"))
			rightStyle = base.Foreground(lipgloss.Color("#00FF00"))
		}
	} else {
		leftStyle = lineContextStyle
		rightStyle = lineContextStyle
		switch ln.Kind {
		case git.KindDeleted:
			leftStyle = lineDeletedStyle
		case git.KindAdded:
			rightStyle = lineAddedStyle
		case git.KindModified:
			leftStyle = lineDeletedStyle
			rightStyle = lineAddedStyle
		}
	}

	return leftStyle.Width(colWidth).Render(leftLine) + " │ " + rightStyle.Width(colWidth).Render(rightLine)
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
