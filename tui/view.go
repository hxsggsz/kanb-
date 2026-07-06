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
		return v
	}
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
		sideWidth = m.width / sidebarDenominator
		if sideWidth < sidebarMinWidth {
			sideWidth = sidebarMinWidth
		}
		if sideWidth > sidebarMaxWidth {
			sideWidth = sidebarMaxWidth
		}
	}

	maxFiles := m.height - (statusBarHeight + borderHeight)
	if maxFiles < 1 {
		maxFiles = 1
	}

	var sb strings.Builder
	for i, f := range m.diffs {
		if i >= maxFiles {
			break
		}
		statusColor := statusColorFor(f.Status)
		label := fmt.Sprintf("%s %s", statusColor.Render(f.Status), f.NewPath)
		if i == m.fileIdx {
			sb.WriteString(sidebarFileSelected.Render("▸ "+label) + "\n")
		} else {
			sb.WriteString(sidebarFile.Render("  "+label) + "\n")
		}
	}
	sidebar := sidebarStyle.Width(sideWidth).Render(sb.String())

	panelWidth := m.width - sideWidth - panelBorderWidth
	if panelWidth < panelMinWidth {
		panelWidth = panelMinWidth
	}

	file := m.diffs[m.fileIdx]
	content := m.renderFile(file, panelWidth)

	statusBar := statusBarStyle.Width(m.width).Render(
		fmt.Sprintf(" %d/%d  •  ↑↓ scroll  •  n/p file  •  g/G top/bottom  •  ? help  •  q quit",
			m.fileIdx+1, len(m.diffs)))

	return fmt.Sprintf("%s%s\n%s", sidebar, content, statusBar)
}

func (m *model) renderFile(f git.FileDiff, width int) string {
	var lines []string
	for _, h := range f.Hunks {
		lines = append(lines, hunkHeaderStyle.Render(h.Header))
		for _, ln := range h.Lines {
			lines = append(lines, formatLine(ln, width))
		}
	}

	start := m.scroll
	end := start + m.visibleLines()
	if start >= len(lines) {
		start = 0
	}
	if end > len(lines) {
		end = len(lines)
	}

	if start >= end {
		return ""
	}

	return strings.Join(lines[start:end], "\n")
}

func formatLine(ln git.Line, width int) string {
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
	style := lineContextStyle

	switch ln.Type {
	case git.LineAdded:
		prefix = "+"
		style = lineAddedStyle
	case git.LineDeleted:
		prefix = "-"
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
	content := " Keybindings\n\n"
	bindings := []struct{ key, desc string }{
		{"↑/k", "Scroll up"},
		{"↓/j", "Scroll down"},
		{"n", "Next file"},
		{"p", "Previous file"},
		{"g", "Go to top"},
		{"G", "Go to bottom"},
		{"?", "Toggle help"},
		{"q/ctrl+c", "Quit"},
	}
	for _, b := range bindings {
		content += fmt.Sprintf("  %-12s %s\n", b.key, b.desc)
	}
	return helpStyle.Render(content)
}
