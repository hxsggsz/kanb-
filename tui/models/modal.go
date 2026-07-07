package models

import (
	"fmt"
	"strings"

	"charm.land/lipgloss/v2"
	"kanba/overlay"
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

func (m *Modal) Render(bgColor, accentColor string) string {
	if len(m.Items) == 0 {
		return ""
	}

	cursorStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(accentColor))
	markStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(accentColor))

	var buf strings.Builder
	buf.WriteString(cursorStyle.Render(fmt.Sprintf(" %s\n\n", m.Title)))
	for _, item := range m.Items {
		cursor := " "
		if item.Key == m.Items[m.Cursor].Key {
			cursor = cursorStyle.Render("◄")
		}
		mark := " "
		if item.Key == m.Selected {
			mark = markStyle.Render("●")
		}
		buf.WriteString(fmt.Sprintf(" %s %s %s\n", cursor, mark, item.Label))
	}
	buf.WriteString(cursorStyle.Render("\n ↑↓ navigate  ↵ select  t/esc/q close"))

	style := lipgloss.NewStyle().
		Border(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color(accentColor)).
		Background(lipgloss.Color(bgColor)).
		Padding(1, 2)

	return style.Render(buf.String())
}

func (m *Modal) Overlay(base string, bgColor, borderColor string, contentLeft, contentWidth int) string {
	if !m.Active {
		return base
	}

	fg := m.Render(bgColor, borderColor)
	fgWidth := lipgloss.Width(fg)
	xOff := contentLeft + max(0, (contentWidth-fgWidth)/2)

	return overlay.Composite(fg, base, overlay.Left, overlay.Center, xOff, 0)
}
