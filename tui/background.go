package tui

import (
	"strings"

	"charm.land/lipgloss/v2"
	"kanba/git"
)

const (
	bgGreen = "48;5;22"
	bgRed   = "48;5;52"
	bgDim   = "48;5;236"
)

func backgroundFor(kind git.LineKind, isLeft bool) string {
	switch kind {
	case git.KindAdded:
		if isLeft {
			return ""
		}
		return bgGreen
	case git.KindDeleted:
		if isLeft {
			return bgRed
		}
		return ""
	case git.KindModified:
		if isLeft {
			return bgRed
		}
		return bgGreen
	default:
		return ""
	}
}

func injectBackground(s, bgCode string) string {
	if bgCode == "" || s == "" {
		return s
	}
	s = strings.ReplaceAll(s, "\x1b[0m", "\x1b[0m\x1b["+bgCode+"m")
	return "\x1b[" + bgCode + "m" + s
}

func injectCursor(s string) string {
	s = strings.ReplaceAll(s, "\x1b[0m", "\x1b[0m\x1b[7m")
	return "\x1b[7m" + s + "\x1b[0m"
}

func padToWidth(s string, width int, bgCode string) string {
	vis := lipgloss.Width(s)
	if vis >= width {
		return s + "\x1b[0m"
	}
	if bgCode == "" {
		return s + strings.Repeat(" ", width-vis) + "\x1b[0m"
	}
	return s + "\x1b[" + bgCode + "m" + strings.Repeat(" ", width-vis) + "\x1b[0m"
}
