package tui

import (
	"fmt"
	"strconv"
	"strings"

	models "kanba/tui/models"

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
	theme := m.currentTheme()
	return lipgloss.NewStyle().
		Foreground(lipgloss.Color(theme.LoadingFg)).
		Padding(2, 4).
		Render(" Loading diffs...")
}

func (m *model) errorView() string {
	theme := m.currentTheme()
	return lipgloss.NewStyle().
		Foreground(lipgloss.Color(theme.ErrorFg)).
		Bold(true).
		Padding(2, 4).
		Render(fmt.Sprintf("Error: %v\n\nPress any key to exit.", m.err))
}

func (m *model) emptyView() string {
	theme := m.currentTheme()
	return lipgloss.NewStyle().
		Foreground(lipgloss.Color(theme.LoadingFg)).
		Padding(2, 4).
		Render(" No changes to show.")
}

func (m *model) diffView() string {
	if len(m.flatLines) == 0 {
		return ""
	}

	theme := m.currentTheme()

	sideWidth := CalculateSideWidth(m.width)
	cursorFileIdx := m.flatLines[m.scroller.CursorLine()].fileIdx

	sidebar := NewSidebar(m.diffs, cursorFileIdx, sideWidth, m.height, theme, m.fileStats)
	sidebarStr := sidebar.Render()

	contentVis := m.visibleLines()
	panelWidth := max(m.width-sideWidth-panelBorderWidth, panelMinWidth)
	content := m.renderContinuous(panelWidth, contentVis)

	f := m.diffs[cursorFileIdx]
	statusBar := NewStatusBar(f.NewPath, cursorFileIdx, len(m.diffs), m.scroller.CursorLine(), len(m.flatLines), m.width, theme)

	body := lipgloss.JoinHorizontal(lipgloss.Top, sidebarStr, content)
	result := fmt.Sprintf("%s\n%s", statusBar.Render(), body)

	theme = m.currentTheme()
	result = m.themeModal.Overlay(result, theme.PanelBg, theme.SidebarSelected, sideWidth, panelWidth)

	return result
}

func (m *model) renderContinuous(width int, vis int) string {
	total := len(m.flatLines)
	if total == 0 {
		return ""
	}
	if vis <= 0 {
		m.scroller.UpdateScroll(total, vis)
		return ""
	}

	theme := m.currentTheme()

	m.scroller.UpdateScroll(total, vis)
	cursorLine := m.scroller.CursorLine()
	hScroll := m.scroller.HScroll()

	start := m.scroller.Scroll()
	end := min(start+vis, total)

	var lines []string
	for gi := start; gi < end; gi++ {
		fl := m.flatLines[gi]
		cursor := gi == cursorLine

		if fl.isHeader {
			lines = append(lines, m.renderFileHeader(fl, width, cursor))
		} else {
			f := m.diffs[fl.fileIdx]
			h := f.Hunks[fl.hunkIdx]
			ln := h.Lines[fl.lineIdx]
			fmtr := defaultFormatters[ln.Kind]

			singlePanel := f.Status == "A"
			colWidth := width
			if !singlePanel {
				colWidth = (width - 3) / 2
			}

			lines = append(lines, renderAlignedLine(fmtr, ln, colWidth, cursor, m.highlighter, f.NewPath, hScroll, singlePanel, theme))
		}
	}

	return strings.Join(lines, "\n")
}

func (m *model) renderFileHeader(fl flatLine, colWidth int, cursor bool) string {
	theme := m.currentTheme()
	f := m.diffs[fl.fileIdx]
	stats := m.fileStats[fl.fileIdx]

	addStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(theme.SidebarAdded))
	delStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(theme.SidebarDeleted))

	var parts []string
	if stats.added > 0 {
		parts = append(parts, addStyle.Render("+"+strconv.Itoa(stats.added)))
	}
	if stats.deleted > 0 {
		parts = append(parts, delStyle.Render("-"+strconv.Itoa(stats.deleted)))
	}

	statsStr := strings.Join(parts, ", ")
	if statsStr != "" {
		statsStr = " (" + statsStr + ")"
	}
	text := fmt.Sprintf(" %s%s", f.NewPath, statsStr)

	style := lipgloss.NewStyle()
	if fl.fileIdx > 0 {
		style = style.
			MarginTop(1).
			BorderTop(true).
			BorderStyle(lipgloss.NormalBorder()).
			BorderForeground(lipgloss.Color(theme.BorderColor))
	}
	style = style.Padding(1, 1).Width(colWidth)
	if cursor {
		style = style.Background(lipgloss.Color(theme.CursorBg))
	}
	return style.Render(text)
}

func statusColorFor(status string, theme models.Theme) lipgloss.Style {
	switch status {
	case "A":
		return lipgloss.NewStyle().Foreground(lipgloss.Color(theme.SidebarAdded))
	case "D":
		return lipgloss.NewStyle().Foreground(lipgloss.Color(theme.SidebarDeleted))
	default:
		return lipgloss.NewStyle().Foreground(lipgloss.Color(theme.SidebarModified))
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
		{"g", "Go to top"},
		{"G", "Go to bottom"},
		{"t", "Open theme selector"},
		{"?", "Toggle help"},
		{"q/ctrl+c", "Quit"},
	}
	for _, b := range bindings {
		fmt.Fprintf(&content, "  %-12s %s\n", b.key, b.desc)
	}
	return helpStyle.Render(content.String())
}
