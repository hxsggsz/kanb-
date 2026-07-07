package tui

import (
	"fmt"
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
	sideWidth := CalculateSideWidth(m.width)

	sidebar := NewSidebar(m.diffs, m.fileIdx, sideWidth, m.height)
	sidebarStr := sidebar.Render()

	contentVis := max(m.height-sidebar.ContentHeight()-2, 1)
	panelWidth := max(m.width-sideWidth-panelBorderWidth, panelMinWidth)

	file := m.diffs[m.fileIdx]
	content := m.renderFile(file, panelWidth, contentVis)

	total := m.totalLines()
	statusBar := NewStatusBar(m.fileIdx, len(m.diffs), m.scroller.CursorLine(), total, m.width)

	body := lipgloss.JoinHorizontal(lipgloss.Top, sidebarStr, content)
	return fmt.Sprintf("%s\n%s", body, statusBar.Render())
}

func (m *model) renderFile(f git.SideBySideDiff, width int, vis int) string {
	total := m.totalLines()
	if total == 0 {
		return ""
	}

	if vis <= 0 {
		m.scroller.UpdateScroll(total, vis)
		return ""
	}

	m.scroller.UpdateScroll(total, vis)

	colWidth := (width - 3) / 2

	// clamp horizontal scroll to file's longest content width
	contentWidth := maxFileContentWidth(f)
	contentAreaWidth := colWidth - (lineNumColWidth + 3)
	m.scroller.ClampHScroll(max(0, contentWidth-contentAreaWidth))

	hScroll := m.scroller.HScroll()
	start := m.scroller.Scroll()
	end := min(start + vis, total)

	var lines []string
	lineIdx := 0
	for _, h := range f.Hunks {
		if lineIdx >= start && lineIdx < end {
			cursor := lineIdx == m.scroller.CursorLine()
			line := hunkHeaderStyle.Render(h.Header)
			if cursor {
				line = lineCursorStyle.Width(width).Render(line)
			}
			lines = append(lines, line)
		}
		lineIdx++

		for _, ln := range h.Lines {
			if lineIdx >= start && lineIdx < end {
				cursor := lineIdx == m.scroller.CursorLine()
				fmtr := defaultFormatters[ln.Kind]
				lines = append(lines, renderAlignedLine(fmtr, ln, colWidth, cursor, m.highlighter, f.NewPath, hScroll))
			}
			lineIdx++
		}

		if lineIdx >= end {
			break
		}
	}

	return strings.Join(lines, "\n")
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

func maxFileContentWidth(f git.SideBySideDiff) int {
	maxWidth := 0
	for _, h := range f.Hunks {
		for _, ln := range h.Lines {
			w := max(len(ln.OldContent), len(ln.NewContent))
			if w > maxWidth {
				maxWidth = w
			}
		}
	}
	return maxWidth
}

func (m *model) helpView() string {
	var content strings.Builder
	content.WriteString(" Keybindings\n\n")
	bindings := []struct{ key, desc string }{
		{"↑/k", "Cursor up"},
		{"↓/j", "Cursor down"},
		{"h/←", "Scroll left 8 cols"},
		{"l/→", "Scroll right 8 cols"},
		{"C-←/C-→", "Scroll 32 cols"},
		{"_", "Go to line start"},
		{"$", "Go to line end"},
		{"n", "Next file"},
		{"p", "Previous file"},
		{"g", "Go to top"},
		{"G", "Go to bottom"},
		{"?", "Toggle help"},
		{"q/ctrl+c", "Quit"},
	}
	for _, b := range bindings {
		fmt.Fprintf(&content, "  %-12s %s\n", b.key, b.desc)
	}
	return helpStyle.Render(content.String())
}
