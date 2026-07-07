package tui

import (
	"fmt"
	"strconv"

	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
	"kanba/git"
)

var emptyStyle = lipgloss.NewStyle()

type LineFormatter interface {
	LeftContent(ln git.AlignedLine) string
	RightContent(ln git.AlignedLine) string
	LeftPrefix(ln git.AlignedLine) string
	RightPrefix(ln git.AlignedLine) string
	LeftStyle(ln git.AlignedLine, cursor bool) lipgloss.Style
	RightStyle(ln git.AlignedLine, cursor bool) lipgloss.Style
}

type contextFormatter struct{}

func (contextFormatter) LeftContent(ln git.AlignedLine) string    { return ln.OldContent }
func (contextFormatter) RightContent(ln git.AlignedLine) string   { return ln.NewContent }
func (contextFormatter) LeftPrefix(git.AlignedLine) string        { return " " }
func (contextFormatter) RightPrefix(git.AlignedLine) string       { return " " }
func (contextFormatter) LeftStyle(git.AlignedLine, bool) lipgloss.Style  { return emptyStyle }
func (contextFormatter) RightStyle(git.AlignedLine, bool) lipgloss.Style { return emptyStyle }

type addedFormatter struct{}

func (addedFormatter) LeftContent(git.AlignedLine) string          { return "" }
func (addedFormatter) RightContent(ln git.AlignedLine) string      { return ln.NewContent }
func (addedFormatter) LeftPrefix(git.AlignedLine) string           { return " " }
func (addedFormatter) RightPrefix(git.AlignedLine) string          { return "+" }
func (addedFormatter) LeftStyle(git.AlignedLine, bool) lipgloss.Style  { return emptyStyle }
func (addedFormatter) RightStyle(git.AlignedLine, bool) lipgloss.Style { return emptyStyle }

type deletedFormatter struct{}

func (deletedFormatter) LeftContent(ln git.AlignedLine) string     { return ln.OldContent }
func (deletedFormatter) RightContent(git.AlignedLine) string       { return "" }
func (deletedFormatter) LeftPrefix(git.AlignedLine) string         { return "-" }
func (deletedFormatter) RightPrefix(git.AlignedLine) string        { return " " }
func (deletedFormatter) LeftStyle(git.AlignedLine, bool) lipgloss.Style  { return emptyStyle }
func (deletedFormatter) RightStyle(git.AlignedLine, bool) lipgloss.Style { return emptyStyle }

type modifiedFormatter struct{}

func (modifiedFormatter) LeftContent(ln git.AlignedLine) string    { return ln.OldContent }
func (modifiedFormatter) RightContent(ln git.AlignedLine) string   { return ln.NewContent }
func (modifiedFormatter) LeftPrefix(git.AlignedLine) string        { return "-" }
func (modifiedFormatter) RightPrefix(git.AlignedLine) string       { return "+" }
func (modifiedFormatter) LeftStyle(git.AlignedLine, bool) lipgloss.Style  { return emptyStyle }
func (modifiedFormatter) RightStyle(git.AlignedLine, bool) lipgloss.Style { return emptyStyle }

var defaultFormatters = NewDefaultFormatters()

func NewDefaultFormatters() map[git.LineKind]LineFormatter {
	return map[git.LineKind]LineFormatter{
		git.KindContext:  contextFormatter{},
		git.KindAdded:    addedFormatter{},
		git.KindDeleted:  deletedFormatter{},
		git.KindModified: modifiedFormatter{},
	}

}

func renderAlignedLine(f LineFormatter, ln git.AlignedLine, colWidth int, cursor bool, sh *SyntaxHighlighter, filePath string, hScroll int) string {
	oldNum := ""
	if ln.OldLineNum > 0 {
		oldNum = strconv.Itoa(ln.OldLineNum)
	}
	newNum := ""
	if ln.NewLineNum > 0 {
		newNum = strconv.Itoa(ln.NewLineNum)
	}

	leftContent := f.LeftContent(ln)
	rightContent := f.RightContent(ln)
	if sh != nil {
		leftContent = sh.Highlight(leftContent, filePath)
		rightContent = sh.Highlight(rightContent, filePath)
	}

	// clip content by hScroll, keeping line numbers and +/- prefix fixed
	// always clip to prevent overflow into the other panel
	contentAreaWidth := colWidth - (lineNumColWidth + 3)
	if contentAreaWidth < 0 {
		contentAreaWidth = 0
	}
	leftContent = ansi.Cut(leftContent, hScroll, hScroll+contentAreaWidth)
	rightContent = ansi.Cut(rightContent, hScroll, hScroll+contentAreaWidth)

	leftLine := fmt.Sprintf("%*s %s %s", lineNumColWidth, oldNum, f.LeftPrefix(ln), leftContent)
	rightLine := fmt.Sprintf("%*s %s %s", lineNumColWidth, newNum, f.RightPrefix(ln), rightContent)

	leftRendered := f.LeftStyle(ln, cursor).Render(leftLine)
	rightRendered := f.RightStyle(ln, cursor).Render(rightLine)

	if sh != nil {
		leftRendered = renderLineWithHighlight(leftRendered, colWidth, ln.Kind, true, cursor)
		rightRendered = renderLineWithHighlight(rightRendered, colWidth, ln.Kind, false, cursor)
	}

	return leftRendered + " │ " + rightRendered
}

func renderLineWithHighlight(s string, colWidth int, kind git.LineKind, isLeft bool, cursor bool) string {
	if cursor {
		s = padToWidth(s, colWidth, "")
		return injectCursor(s)
	}
	bgCode := backgroundFor(kind, isLeft)
	if bgCode != "" {
		s = injectBackground(s, bgCode)
	}
	return padToWidth(s, colWidth, bgCode)
}
