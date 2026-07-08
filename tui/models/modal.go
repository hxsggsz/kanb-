package models

import (
	"strings"

	"charm.land/lipgloss/v2"
	"kanba/tui/overlay"
)

type ModalItem struct {
	Key   string
	Label string
}

type Modal struct {
	Title    string
	Items    []ModalItem
	Cursor   int
	Selected string
	Active   bool
}

func NewModal(title string, items []ModalItem) *Modal {
	return &Modal{
		Title:  title,
		Items:  items,
		Cursor: 0,
	}
}

func (m *Modal) MoveUp() {
	m.Cursor--
	if m.Cursor < 0 {
		m.Cursor = len(m.Items) - 1
	}
}

func (m *Modal) MoveDown() {
	m.Cursor++
	if m.Cursor >= len(m.Items) {
		m.Cursor = 0
	}
}

func (m *Modal) Select() string {
	if m.Cursor >= 0 && m.Cursor < len(m.Items) {
		m.Selected = m.Items[m.Cursor].Key
	}
	m.Active = false
	return m.Selected
}

func (m *Modal) Close() {
	m.Active = false
}

func (m *Modal) SyncCursor(key string) {
	for i, item := range m.Items {
		if item.Key == key {
			m.Cursor = i
			return
		}
	}
}

func (m *Modal) Render(bgColor, accentColor, fgColor string) string {
	if len(m.Items) == 0 {
		return ""
	}

	bg := lipgloss.Color(bgColor)
	accent := lipgloss.Color(accentColor)

	indicatorStyle := lipgloss.NewStyle().Foreground(accent).Background(bg)
	labelStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(fgColor)).Background(bg)

	var buf strings.Builder

	title := " " + m.Title
	buf.WriteString(indicatorStyle.Render(title))
	buf.WriteString("\n\n")

	for _, item := range m.Items {
		hasCursor := item.Key == m.Items[m.Cursor].Key
		hasMark := item.Key == m.Selected

		var indicator string
		if hasCursor && hasMark {
			indicator = indicatorStyle.Render(" ◄ ● ")
		} else if hasCursor {
			indicator = indicatorStyle.Render(" ◄   ")
		} else if hasMark {
			indicator = indicatorStyle.Render("   ● ")
		} else {
			indicator = "     "
		}

		label := labelStyle.Render(item.Label)
		buf.WriteString(indicator + label + "\n")
	}

	buf.WriteString("\n")
	buf.WriteString(indicatorStyle.Render(" ↑↓ navigate  ↵ select  t/esc/q close"))

	content := buf.String()

	maxW := 0
	for _, line := range strings.Split(content, "\n") {
		w := lipgloss.Width(line)
		if w > maxW {
			maxW = w
		}
	}

	style := lipgloss.NewStyle().
		Background(bg).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(accent).
		BorderBackground(bg).
		Padding(1, 2).
		Width(maxW + 6)

	return style.Render(content)
}

func (m *Modal) Overlay(base string, bgColor, borderColor, fgColor string, contentLeft, contentWidth int) string {
	if !m.Active {
		return base
	}

	fg := m.Render(bgColor, borderColor, fgColor)
	fgWidth := lipgloss.Width(fg)
	xOff := contentLeft + max(0, (contentWidth-fgWidth)/2)

	return overlay.Composite(fg, base, overlay.Left, overlay.Center, xOff, 0)
}
