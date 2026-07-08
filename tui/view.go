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
	theme := m.currentTheme()
	return lipgloss.NewStyle().
		Background(lipgloss.Color(theme.PanelBg)).
		Foreground(lipgloss.Color(theme.LoadingFg)).
		Padding(2, 4).
		Render(" Loading diffs...")
}

func (m *model) errorView() string {
	theme := m.currentTheme()
	return lipgloss.NewStyle().
		Background(lipgloss.Color(theme.PanelBg)).
		Foreground(lipgloss.Color(theme.ErrorFg)).
		Bold(true).
		Padding(2, 4).
		Render(fmt.Sprintf("Error: %v\n\nPress any key to exit.", m.err))
}

func (m *model) emptyView() string {
	theme := m.currentTheme()
	return lipgloss.NewStyle().
		Background(lipgloss.Color(theme.PanelBg)).
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
	result = lipgloss.NewStyle().Background(lipgloss.Color(theme.PanelBg)).Render(result)
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

	bgColor := theme.PanelBg
	if cursor {
		bgColor = theme.CursorBgFor(bgColor)
	}

	bg := lipgloss.NewStyle().Background(lipgloss.Color(bgColor))
	normalStyle := bg.Foreground(lipgloss.Color(theme.ContextFg))
	addStyle := bg.Foreground(lipgloss.Color(theme.SidebarAdded))
	delStyle := bg.Foreground(lipgloss.Color(theme.SidebarDeleted))

	var segs []string
	segs = append(segs, normalStyle.Render(" "+f.NewPath))

	if stats.added > 0 || stats.deleted > 0 {
		segs = append(segs, normalStyle.Render(" ("))
		var statSegs []string
		if stats.added > 0 {
			statSegs = append(statSegs, addStyle.Render("+"+strconv.Itoa(stats.added)))
		}
		if stats.deleted > 0 {
			statSegs = append(statSegs, delStyle.Render("-"+strconv.Itoa(stats.deleted)))
		}
		segs = append(segs, strings.Join(statSegs, normalStyle.Render(", ")))
		segs = append(segs, normalStyle.Render(")"))
	}

	text := strings.Join(segs, "")

	style := lipgloss.NewStyle().
		Background(lipgloss.Color(bgColor)).
		MarginBackground(lipgloss.Color(bgColor)).
		Padding(1, 1).
		Width(colWidth)
	if fl.fileIdx > 0 {
		style = style.MarginTop(1)
	}
	return style.Render(text)
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
	theme := m.currentTheme()
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
	return helpStyle.
		Background(lipgloss.Color(theme.PanelBg)).
		Foreground(lipgloss.Color(theme.ContextFg)).
		Render(content.String())
}
