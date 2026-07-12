package diff

import (
	"fmt"
	"strconv"
	"strings"

	models "kanba/tui/models"

	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
	"kanba/git"
)

const LineNumColWidth = 4

var emptyStyle = lipgloss.NewStyle()

type LineFormatter interface {
	LeftContent(ln git.AlignedLine) string
	RightContent(ln git.AlignedLine) string
	LeftPrefix(ln git.AlignedLine) string
	RightPrefix(ln git.AlignedLine) string
	LeftStyle(ln git.AlignedLine) lipgloss.Style
	RightStyle(ln git.AlignedLine) lipgloss.Style
}

type contextFormatter struct{}

func (contextFormatter) LeftContent(ln git.AlignedLine) string    { return ln.OldContent }
func (contextFormatter) RightContent(ln git.AlignedLine) string   { return ln.NewContent }
func (contextFormatter) LeftPrefix(git.AlignedLine) string        { return " " }
func (contextFormatter) RightPrefix(git.AlignedLine) string       { return " " }
func (contextFormatter) LeftStyle(git.AlignedLine) lipgloss.Style  { return emptyStyle }
func (contextFormatter) RightStyle(git.AlignedLine) lipgloss.Style { return emptyStyle }

type addedFormatter struct{}

func (addedFormatter) LeftContent(git.AlignedLine) string          { return "" }
func (addedFormatter) RightContent(ln git.AlignedLine) string      { return ln.NewContent }
func (addedFormatter) LeftPrefix(git.AlignedLine) string           { return " " }
func (addedFormatter) RightPrefix(git.AlignedLine) string          { return "+" }
func (addedFormatter) LeftStyle(git.AlignedLine) lipgloss.Style  { return emptyStyle }
func (addedFormatter) RightStyle(git.AlignedLine) lipgloss.Style { return emptyStyle }

type deletedFormatter struct{}

func (deletedFormatter) LeftContent(ln git.AlignedLine) string     { return ln.OldContent }
func (deletedFormatter) RightContent(git.AlignedLine) string       { return "" }
func (deletedFormatter) LeftPrefix(git.AlignedLine) string         { return "-" }
func (deletedFormatter) RightPrefix(git.AlignedLine) string        { return " " }
func (deletedFormatter) LeftStyle(git.AlignedLine) lipgloss.Style  { return emptyStyle }
func (deletedFormatter) RightStyle(git.AlignedLine) lipgloss.Style { return emptyStyle }

type modifiedFormatter struct{}

func (modifiedFormatter) LeftContent(ln git.AlignedLine) string    { return ln.OldContent }
func (modifiedFormatter) RightContent(ln git.AlignedLine) string   { return ln.NewContent }
func (modifiedFormatter) LeftPrefix(git.AlignedLine) string        { return "-" }
func (modifiedFormatter) RightPrefix(git.AlignedLine) string       { return "+" }
func (modifiedFormatter) LeftStyle(git.AlignedLine) lipgloss.Style  { return emptyStyle }
func (modifiedFormatter) RightStyle(git.AlignedLine) lipgloss.Style { return emptyStyle }

var DefaultFormatters = NewDefaultFormatters()

func NewDefaultFormatters() map[git.LineKind]LineFormatter {
	return map[git.LineKind]LineFormatter{
		git.KindContext:  contextFormatter{},
		git.KindAdded:    addedFormatter{},
		git.KindDeleted:  deletedFormatter{},
		git.KindModified: modifiedFormatter{},
	}

}

func RenderAlignedLine(f LineFormatter, ln git.AlignedLine, colWidth int, sh *SyntaxHighlighter, filePath string, hScroll int, singlePanel bool, theme models.Theme) string {
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

	contentAreaWidth := colWidth - (LineNumColWidth + 3)
	if contentAreaWidth < 0 {
		contentAreaWidth = 0
	}
	leftContent = ansi.Cut(leftContent, hScroll, hScroll+contentAreaWidth)
	rightContent = ansi.Cut(rightContent, hScroll, hScroll+contentAreaWidth)

	if singlePanel {
		prefix := fmt.Sprintf("%*s %s ", LineNumColWidth, newNum, f.RightPrefix(ln))
		return renderStyledLine(prefix, rightContent, colWidth, ln.Kind, false, sh, filePath, theme)
	}

	leftPrefix := fmt.Sprintf("%*s %s ", LineNumColWidth, oldNum, f.LeftPrefix(ln))
	rightPrefix := fmt.Sprintf("%*s %s ", LineNumColWidth, newNum, f.RightPrefix(ln))

	leftRendered := renderStyledLine(leftPrefix, leftContent, colWidth, ln.Kind, true, sh, filePath, theme)
	rightRendered := renderStyledLine(rightPrefix, rightContent, colWidth, ln.Kind, false, sh, filePath, theme)

	return leftRendered + rightRendered
}

func renderStyledLine(prefix, content string, width int, kind git.LineKind, isLeft bool, sh *SyntaxHighlighter, filePath string, theme models.Theme) string {
	bgColor := theme.BgFor(kind, isLeft)
	if bgColor == "" {
		bgColor = theme.PanelBg
	}

	numBg := bgColor
	if kind == git.KindContext {
		numBg = theme.LineNumberBg
	}

	numStyle := lipgloss.NewStyle()
	if fg := theme.LineNumFg(kind, isLeft); fg != "" {
		numStyle = numStyle.Foreground(lipgloss.Color(fg))
	}
	numStyle = numStyle.Background(lipgloss.Color(numBg))

	prefixRendered := numStyle.Render(prefix)

	var contentRendered string
	if sh != nil {
		baseStyle := lipgloss.NewStyle().Background(lipgloss.Color(bgColor))
		contentRendered = sh.HighlightWithStyle(content, filePath, baseStyle, theme)
	} else {
		contentRendered = content
	}

	styled := prefixRendered + contentRendered
	vis := lipgloss.Width(styled)
	if vis > width {
		styled = ansi.Truncate(styled, width, "")
		vis = width
	}
	if vis < width {
		padStyle := lipgloss.NewStyle().Background(lipgloss.Color(bgColor))
		styled += padStyle.Render(strings.Repeat(" ", width-vis))
	}

	return styled
}
