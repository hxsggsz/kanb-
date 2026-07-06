package tui

import (
	"fmt"
	"strconv"

	"charm.land/lipgloss/v2"
	"kanba/git"
)

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
func (contextFormatter) LeftStyle(_ git.AlignedLine, cursor bool) lipgloss.Style {
	if cursor {
		return lipgloss.NewStyle().Background(lipgloss.Color("#444444"))
	}
	return lineContextStyle
}
func (contextFormatter) RightStyle(_ git.AlignedLine, cursor bool) lipgloss.Style {
	if cursor {
		return lipgloss.NewStyle().Background(lipgloss.Color("#444444"))
	}
	return lineContextStyle
}

type addedFormatter struct{}

func (addedFormatter) LeftContent(git.AlignedLine) string          { return "" }
func (addedFormatter) RightContent(ln git.AlignedLine) string      { return ln.NewContent }
func (addedFormatter) LeftPrefix(git.AlignedLine) string           { return " " }
func (addedFormatter) RightPrefix(git.AlignedLine) string          { return "+" }
func (addedFormatter) LeftStyle(_ git.AlignedLine, cursor bool) lipgloss.Style {
	if cursor {
		return lipgloss.NewStyle().Background(lipgloss.Color("#444444"))
	}
	return lineContextStyle
}
func (addedFormatter) RightStyle(_ git.AlignedLine, cursor bool) lipgloss.Style {
	if cursor {
		return lipgloss.NewStyle().Background(lipgloss.Color("#444444")).Foreground(lipgloss.Color("#00FF00"))
	}
	return lineAddedStyle
}

type deletedFormatter struct{}

func (deletedFormatter) LeftContent(ln git.AlignedLine) string     { return ln.OldContent }
func (deletedFormatter) RightContent(git.AlignedLine) string       { return "" }
func (deletedFormatter) LeftPrefix(git.AlignedLine) string         { return "-" }
func (deletedFormatter) RightPrefix(git.AlignedLine) string        { return " " }
func (deletedFormatter) LeftStyle(_ git.AlignedLine, cursor bool) lipgloss.Style {
	if cursor {
		return lipgloss.NewStyle().Background(lipgloss.Color("#444444")).Foreground(lipgloss.Color("#FF0000"))
	}
	return lineDeletedStyle
}
func (deletedFormatter) RightStyle(_ git.AlignedLine, cursor bool) lipgloss.Style {
	if cursor {
		return lipgloss.NewStyle().Background(lipgloss.Color("#444444"))
	}
	return lineContextStyle
}

type modifiedFormatter struct{}

func (modifiedFormatter) LeftContent(ln git.AlignedLine) string    { return ln.OldContent }
func (modifiedFormatter) RightContent(ln git.AlignedLine) string   { return ln.NewContent }
func (modifiedFormatter) LeftPrefix(git.AlignedLine) string        { return "-" }
func (modifiedFormatter) RightPrefix(git.AlignedLine) string       { return "+" }
func (modifiedFormatter) LeftStyle(_ git.AlignedLine, cursor bool) lipgloss.Style {
	if cursor {
		return lipgloss.NewStyle().Background(lipgloss.Color("#444444")).Foreground(lipgloss.Color("#FF0000"))
	}
	return lineDeletedStyle
}
func (modifiedFormatter) RightStyle(_ git.AlignedLine, cursor bool) lipgloss.Style {
	if cursor {
		return lipgloss.NewStyle().Background(lipgloss.Color("#444444")).Foreground(lipgloss.Color("#00FF00"))
	}
	return lineAddedStyle
}

var defaultFormatters = NewDefaultFormatters()

func NewDefaultFormatters() map[git.LineKind]LineFormatter {
	return map[git.LineKind]LineFormatter{
		git.KindContext:  contextFormatter{},
		git.KindAdded:    addedFormatter{},
		git.KindDeleted:  deletedFormatter{},
		git.KindModified: modifiedFormatter{},
	}
}

func renderAlignedLine(f LineFormatter, ln git.AlignedLine, colWidth int, cursor bool) string {
	oldNum := ""
	if ln.OldLineNum > 0 {
		oldNum = strconv.Itoa(ln.OldLineNum)
	}
	newNum := ""
	if ln.NewLineNum > 0 {
		newNum = strconv.Itoa(ln.NewLineNum)
	}

	leftLine := fmt.Sprintf("%*s %s %s", lineNumColWidth, oldNum, f.LeftPrefix(ln), f.LeftContent(ln))
	rightLine := fmt.Sprintf("%*s %s %s", lineNumColWidth, newNum, f.RightPrefix(ln), f.RightContent(ln))

	return f.LeftStyle(ln, cursor).Width(colWidth).Render(leftLine) +
		" │ " +
		f.RightStyle(ln, cursor).Width(colWidth).Render(rightLine)
}
