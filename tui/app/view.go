package app

import (
	"fmt"
	"strconv"
	"strings"

	"kanba/tui/diff"
	"kanba/tui/models"
	"kanba/tui/overlay"
	"kanba/tui/widget"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
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

	if m.screen == screenDiff {
		v.SetContent(m.diffView())
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

func (m *Model) diffView() string {
	if len(m.flatLines) == 0 {
		return ""
	}

	theme := m.CurrentTheme()

	sideWidth := widget.CalculateSideWidth(m.width)
	cursorFileIdx := m.flatLines[m.scroller.CursorLine()].FileIdx

	sidebar := widget.NewSidebar(m.diffs, cursorFileIdx, sideWidth, m.height, theme, m.fileStats)
	sidebarStr := sidebar.Render()

	contentVis := m.VisibleLines()
	panelWidth := max(m.width-sideWidth-panelBorderWidth, panelMinWidth)
	content := m.renderContinuous(panelWidth, contentVis)

	f := m.diffs[cursorFileIdx]
	statusBar := widget.NewStatusBar(f.NewPath, cursorFileIdx, len(m.diffs), m.scroller.CursorLine(), len(m.flatLines), m.width, theme)

	body := lipgloss.JoinHorizontal(lipgloss.Top, sidebarStr, content)
	result := fmt.Sprintf("%s\n%s", statusBar.Render(), body)

	theme = m.CurrentTheme()
	result = lipgloss.NewStyle().Background(lipgloss.Color(theme.PanelBg)).Render(result)
	result = m.themeModal.Overlay(result, theme.PanelBg, theme.SidebarSelected, theme.ContextFg, sideWidth, panelWidth)

	if m.helpActive {
		result = m.helpOverlay(result, theme, sideWidth, panelWidth)
	}

	return result
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

	var lines []string
	for gi := start; gi < end; gi++ {
		fl := m.flatLines[gi]
		cursor := gi == cursorLine

		if fl.IsHeader {
			lines = append(lines, m.renderFileHeader(fl, width, cursor))
		} else {
			f := m.diffs[fl.FileIdx]
			h := f.Hunks[fl.HunkIdx]
			ln := h.Lines[fl.LineIdx]
			fmtr := diff.DefaultFormatters[ln.Kind]

			singlePanel := f.Status == "A"
			colWidth := width
			if !singlePanel {
				colWidth = (width - 3) / 2
			}

			lines = append(lines, diff.RenderAlignedLine(fmtr, ln, colWidth, cursor, m.highlighter, f.NewPath, hScroll, singlePanel, theme))
		}
	}

	return strings.Join(lines, "\n")
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
