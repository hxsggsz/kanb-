package tui

import (
	"strconv"
	"strings"

	models "kanba/tui/models"

	"charm.land/lipgloss/v2"
	"kanba/git"
)

type Panel struct {
	width       int
	theme       models.Theme
	diffs       []git.SideBySideDiff
	flatLines   []flatLine
	fileStats   []fileStat
	scroller    *Scroller
	highlighter *SyntaxHighlighter
}

func NewPanel() *Panel {
	return &Panel{}
}

func (p *Panel) Width(w int) *Panel                { p.width = w; return p }
func (p *Panel) Theme(t models.Theme) *Panel        { p.theme = t; return p }
func (p *Panel) Diffs(d []git.SideBySideDiff) *Panel { p.diffs = d; return p }
func (p *Panel) FlatLines(f []flatLine) *Panel      { p.flatLines = f; return p }
func (p *Panel) FileStats(s []fileStat) *Panel      { p.fileStats = s; return p }
func (p *Panel) Scroller(s *Scroller) *Panel        { p.scroller = s; return p }
func (p *Panel) Highlighter(h *SyntaxHighlighter) *Panel { p.highlighter = h; return p }

func (p *Panel) Render(vis int) string {
	total := len(p.flatLines)
	if total == 0 {
		return ""
	}
	if vis <= 0 {
		p.scroller.UpdateScroll(total, vis)
		return ""
	}

	p.scroller.UpdateScroll(total, vis)
	cursorLine := p.scroller.CursorLine()
	hScroll := p.scroller.HScroll()

	start := p.scroller.Scroll()
	end := min(start+vis, total)

	innerWidth := max(p.width-2, 0)

	type fileBlock struct {
		lines []string
	}
	var blocks []fileBlock
	var cur []string
	curFile := -1

	flush := func() {
		if cur != nil {
			blocks = append(blocks, fileBlock{lines: cur})
			cur = nil
		}
	}

	for gi := start; gi < end; gi++ {
		fl := p.flatLines[gi]
		cursor := gi == cursorLine

		if fl.fileIdx != curFile {
			flush()
			curFile = fl.fileIdx
		}

		var line string
		if fl.isHeader {
			line = p.renderFileHeader(fl, innerWidth, cursor)
		} else {
			f := p.diffs[fl.fileIdx]
			h := f.Hunks[fl.hunkIdx]
			ln := h.Lines[fl.lineIdx]
			fmtr := defaultFormatters[ln.Kind]

			singlePanel := f.Status == "A"
			colWidth := innerWidth
			if !singlePanel {
				colWidth = (innerWidth - 3) / 2
			}

			line = renderAlignedLine(fmtr, ln, colWidth, cursor, p.highlighter, f.NewPath, hScroll, singlePanel, p.theme)
		}

		cur = append(cur, line)
	}
	flush()

	var rendered []string
	borderStyle := lipgloss.NewStyle().
		Background(lipgloss.Color(p.theme.PanelBg)).
		Width(p.width)

	for _, b := range blocks {
		content := strings.Join(b.lines, "\n")
		rendered = append(rendered, borderStyle.Render(content))
	}

	return strings.Join(rendered, "\n")
}

func (p *Panel) renderFileHeader(fl flatLine, colWidth int, cursor bool) string {
	f := p.diffs[fl.fileIdx]
	stats := p.fileStats[fl.fileIdx]

	bgColor := p.theme.PanelHeaderBg
	if cursor {
		bgColor = p.theme.CursorBgFor(bgColor)
	}

	bg := lipgloss.NewStyle().Background(lipgloss.Color(bgColor))
	normalStyle := bg.Foreground(lipgloss.Color(p.theme.ContextFg))
	addStyle := bg.Foreground(lipgloss.Color(p.theme.SidebarAdded))
	delStyle := bg.Foreground(lipgloss.Color(p.theme.SidebarDeleted))

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
	return style.Render(text)
}
