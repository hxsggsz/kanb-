package app

import (
	"fmt"
	"strconv"
	"strings"

	"kanba/tui/diff"
	"kanba/tui/models"
	"kanba/tui/overlay"
	"kanba/tui/selection"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
	"kanba/git"
)

func (m *Model) View() tea.View {
	v := tea.NewView("")
	v.AltScreen = true
	v.MouseMode = tea.MouseModeCellMotion
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

	if m.activeMode != nil {
		v.SetContent(m.activeMode.Render(m))
	}

	return v
}

func (m *Model) loadingView() string {
	theme := m.CurrentTheme()
	return lipgloss.NewStyle().
		Background(lipgloss.Color(theme.PanelBg)).
		Foreground(lipgloss.Color(theme.LoadingFg)).
		Padding(2, 4).
		Render(" Loading diffs...")
}

func (m *Model) errorView() string {
	theme := m.CurrentTheme()
	return lipgloss.NewStyle().
		Background(lipgloss.Color(theme.PanelBg)).
		Foreground(lipgloss.Color(theme.ErrorFg)).
		Bold(true).
		Padding(2, 4).
		Render(fmt.Sprintf("Error: %v\n\nPress any key to exit.", m.err))
}

func (m *Model) emptyView() string {
	theme := m.CurrentTheme()
	return lipgloss.NewStyle().
		Background(lipgloss.Color(theme.PanelBg)).
		Foreground(lipgloss.Color(theme.LoadingFg)).
		Padding(2, 4).
		Render(" No changes to show.")
}

func (m *Model) renderContinuous(width int, vis int) string {
	total := len(m.flatLines)
	if total == 0 {
		return ""
	}
	if vis <= 0 {
		m.scroller.UpdateScroll(total, vis)
		return ""
	}

	theme := m.CurrentTheme()

	m.scroller.UpdateScroll(total, vis)
	cursorLine := m.scroller.CursorLine()
	hScroll := m.scroller.HScroll()

	start := m.scroller.Scroll()
	end := min(start+vis, total)

	padStyle := lipgloss.NewStyle().Background(lipgloss.Color(theme.PanelBg))

	var lines []string
	for gi := start; gi < end; gi++ {
		fl := m.flatLines[gi]
		cursor := gi == cursorLine

		var line string
		if fl.IsHeader {
			line = m.renderFileHeader(fl, width, cursor)
		} else {
			f := m.diffs[fl.FileIdx]
			h := f.Hunks[fl.HunkIdx]
			ln := h.Lines[fl.LineIdx]
			fmtr := diff.DefaultFormatters[ln.Kind]

			singlePanel := f.Status == "A"
			colWidth := width
			if !singlePanel {
				colWidth = width / 2
			}

			line = diff.RenderAlignedLine(fmtr, ln, colWidth, cursor, m.highlighter, f.NewPath, hScroll, singlePanel, theme)

			if m.selection != nil {
				sel := m.selection.CurrentSelection()
				if sel != nil && !sel.Range.IsEmpty() {
					line = m.applySelectionHighlight(line, gi, sel, colWidth, singlePanel)
				}
			}
		}

		if w := lipgloss.Width(line); w < width {
			line += padStyle.Render(strings.Repeat(" ", width-w))
		}
		lines = append(lines, line)
	}

	return strings.Join(lines, "\n")
}

func (m *Model) applySelectionHighlight(line string, flatIdx int, sel *selection.Selection, colWidth int, singlePanel bool) string {
	normalized := sel.Range.Normalized()

	isBeforeSelection := flatIdx < normalized.StartLine
	isAfterSelection := flatIdx > normalized.EndLine
	if isBeforeSelection || isAfterSelection {
		return line
	}

	startCol, endCol := 0, colWidth
	isFirstLine := flatIdx == normalized.StartLine
	isLastLine := flatIdx == normalized.EndLine
	if isFirstLine {
		startCol = normalized.StartCol
	}
	if isLastLine {
		endCol = normalized.EndCol
	}

	if !singlePanel && sel.Panel == selection.PanelRight {
		startCol += colWidth
		endCol += colWidth
	}

	theme := m.CurrentTheme()
	return highlightColumns(line, startCol, endCol, theme.SelectionBg, theme.SelectionFg)
}

func highlightColumns(line string, startCol, endCol int, bgColor, fgColor string) string {
	if bgColor == "" || startCol >= endCol {
		return line
	}

	visWidth := ansi.StringWidth(line)
	if startCol >= visWidth {
		return line
	}
	if endCol > visWidth {
		endCol = visWidth
	}

	before := ansi.Cut(line, 0, startCol)
	selected := ansi.Cut(line, startCol, endCol)
	after := ansi.Cut(line, endCol, visWidth)

	selStyle := lipgloss.NewStyle().Background(lipgloss.Color(bgColor))
	if fgColor != "" {
		selStyle = selStyle.Foreground(lipgloss.Color(fgColor))
	}
	return before + selStyle.Render(selected) + after
}

func (m *Model) renderFileHeader(fl diff.FlatLine, colWidth int, cursor bool) string {
	theme := m.CurrentTheme()
	f := m.diffs[fl.FileIdx]
	stats := m.fileStats[fl.FileIdx]

	bgColor := theme.PanelHeaderBg
	if cursor {
		bgColor = theme.CursorBgFor(bgColor)
	}

	bg := lipgloss.NewStyle().Background(lipgloss.Color(bgColor))
	normalStyle := bg.Foreground(lipgloss.Color(theme.ContextFg))
	addStyle := bg.Foreground(lipgloss.Color(theme.SidebarAdded))
	delStyle := bg.Foreground(lipgloss.Color(theme.SidebarDeleted))

	var segs []string
	segs = append(segs, normalStyle.Render(" "+f.NewPath))

	if stats.Added > 0 || stats.Deleted > 0 {
		segs = append(segs, normalStyle.Render(" ("))
		var statSegs []string
		if stats.Added > 0 {
			statSegs = append(statSegs, addStyle.Render("+"+strconv.Itoa(stats.Added)))
		}
		if stats.Deleted > 0 {
			statSegs = append(statSegs, delStyle.Render("-"+strconv.Itoa(stats.Deleted)))
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
	if fl.FileIdx > 0 {
		style = style.
			Border(lipgloss.NormalBorder(), true, false, false, false).
			BorderForeground(lipgloss.Color(theme.SidebarDir)).
			BorderBackground(lipgloss.Color(bgColor))
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

func (m *Model) helpOverlay(base string, theme models.Theme, sideWidth, panelWidth int) string {
	fg := m.helpContent(theme)
	fgWidth := lipgloss.Width(fg)
	xOff := sideWidth + max(0, (panelWidth-fgWidth)/2)
	return overlay.Composite(fg, base, overlay.Left, overlay.Center, xOff, 0)
}

func (m *Model) helpContent(theme models.Theme) string {
	bg := lipgloss.Color(theme.PanelBg)
	accent := lipgloss.Color(theme.SidebarSelected)

	accentStyle := lipgloss.NewStyle().Foreground(accent).Background(bg)

	var buf strings.Builder
	buf.WriteString(accentStyle.Render(" Keybindings"))
	buf.WriteString("\n\n")
	bindings := []struct{ key, desc string }{
		{"\u2191/k", "Cursor up"},
		{"\u2193/j", "Cursor down"},
		{"h/\u2190", "Scroll left 8 cols"},
		{"l/\u2192", "Scroll right 8 cols"},
		{"C-\u2190/C-\u2192", "Scroll 32 cols"},
		{"_", "Go to line start"},
		{"$", "Go to line end"},
		{"g", "Go to top"},
		{"G", "Go to bottom"},
		{"t", "Open theme selector"},
		{"?", "Toggle help"},
		{"q/ctrl+c", "Quit"},
	}
	for _, b := range bindings {
		fmt.Fprintf(&buf, "  %-12s %s\n", b.key, b.desc)
	}
	buf.WriteString("\n")
	buf.WriteString(accentStyle.Render(" ?/esc close"))

	content := buf.String()

	maxW := 0
	for _, line := range strings.Split(content, "\n") {
		w := lipgloss.Width(line)
		if w > maxW {
			maxW = w
		}
	}

	return lipgloss.NewStyle().
		Background(bg).
		Foreground(lipgloss.Color(theme.ContextFg)).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(accent).
		BorderBackground(bg).
		Padding(1, 2).
		Width(maxW + 6).
		Render(content)
}
