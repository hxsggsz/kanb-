package app

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"kanba/tui/diff"
	"kanba/tui/models"
	"kanba/tui/overlay"

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
	hScroll := m.scroller.HScroll()

	start := m.scroller.Scroll()
	end := min(start+vis, total)

	padStyle := lipgloss.NewStyle().Background(lipgloss.Color(theme.PanelBg))

	selHighlighter := m.buildSelectionHighlighter(width)

	var lines []string
	var selectedTextParts []string

	for gi := start; gi < end; gi++ {
		fl := m.flatLines[gi]

		line := m.renderLine(fl, width, hScroll, selHighlighter, gi, theme)
		selectedTextParts = m.accumulateSelectedText(selectedTextParts, line, gi, selHighlighter, theme)

		if w := lipgloss.Width(line); w < width {
			line += padStyle.Render(strings.Repeat(" ", width-w))
		}
		lines = append(lines, line)
	}

	m.selectedText = strings.Join(selectedTextParts, "\n")
	return strings.Join(lines, "\n")
}

func (m *Model) buildSelectionHighlighter(width int) *SelectionHighlighter {
	if m.selection == nil {
		return nil
	}

	sel := m.selection.CurrentSelection()
	if sel == nil {
		return nil
	}
	if sel.Range.IsEmpty() {
		return nil
	}

	return NewSelectionHighlighter(sel, width/2)
}

func (m *Model) renderLine(fl diff.FlatLine, width int, hScroll int, selHighlighter *SelectionHighlighter, gi int, theme models.Theme) string {
	if fl.IsHeader {
		return m.renderFileHeader(fl, width)
	}

	f := m.diffs[fl.FileIdx]
	h := f.Hunks[fl.HunkIdx]
	ln := h.Lines[fl.LineIdx]
	fmtr := diff.DefaultFormatters[ln.Kind]

	singlePanel := f.Status == "A"
	colWidth := width
	if !singlePanel {
		colWidth = width / 2
	}

	line := diff.RenderAlignedLine(fmtr, ln, colWidth, m.highlighter, f.NewPath, hScroll, singlePanel, theme)

	if selHighlighter == nil {
		return line
	}

	highlighted, _ := selHighlighter.ProcessLine(line, gi, theme)
	return highlighted
}

func (m *Model) accumulateSelectedText(parts []string, line string, gi int, selHighlighter *SelectionHighlighter, theme models.Theme) []string {
	if selHighlighter == nil {
		return parts
	}

	_, plainText := selHighlighter.ProcessLine(line, gi, theme)
	if plainText == "" {
		return parts
	}

	return append(parts, plainText)
}

var (
	sgrRegex   = regexp.MustCompile(`\x1b\[([\d;]*)m`)
	resetRegex = regexp.MustCompile(`\x1b\[m`)
)

func highlightColumns(line string, startCol, endCol int, bgColor string) string {
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

	selected = stripBackgrounds(selected)
	selected = resetRegex.ReplaceAllString(selected, "")
	selStyle := lipgloss.NewStyle().Background(lipgloss.Color(bgColor))
	return before + selStyle.Render(selected) + after
}

func stripBackgrounds(s string) string {
	return sgrRegex.ReplaceAllStringFunc(s, func(match string) string {
		params := match[2 : len(match)-1]
		parts := strings.Split(params, ";")
		var out []string
		i := 0
		for i < len(parts) {
			if parts[i] == "48" && i+1 < len(parts) {
				if parts[i+1] == "5" && i+2 < len(parts) {
					i += 3
					continue
				}
				if parts[i+1] == "2" && i+4 < len(parts) {
					i += 5
					continue
				}
			}
			out = append(out, parts[i])
			i++
		}
		if len(out) == 0 {
			return "\x1b[m"
		}
		return "\x1b[" + strings.Join(out, ";") + "m"
	})
}

func (m *Model) renderFileHeader(fl diff.FlatLine, colWidth int) string {
	theme := m.CurrentTheme()
	f := m.diffs[fl.FileIdx]
	stats := m.fileStats[fl.FileIdx]

	bgColor := theme.PanelHeaderBg

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
		{"\u2191/k", "Scroll up"},
		{"\u2193/j", "Scroll down"},
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
