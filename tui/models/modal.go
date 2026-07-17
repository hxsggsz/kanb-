package models

import (
	"fmt"
	"strings"

	"charm.land/lipgloss/v2"
	"kanba/tui/overlay"
)

const maxVisibleItems = 15

// ListStartLine is the Y offset within the rendered modal output where item
// lines begin (1 top border + 1 top padding + 4 content header lines).
const ListStartLine = 6

type ModalItem struct {
	Key   string
	Label string
}

type Modal struct {
	Title        string
	Items        []ModalItem
	Filtered     []ModalItem
	FilterQuery  string
	IsFocused    bool
	Cursor       int
	ScrollOffset int
	Selected     string
	Active       bool
}

func NewModal(title string, items []ModalItem) *Modal {
	m := &Modal{
		Title:     title,
		Items:     items,
		Cursor:    0,
		IsFocused: true,
	}
	m.Filtered = items
	return m
}

func (m *Modal) VisibleCount() int {
	n := len(m.Filtered)
	if n < maxVisibleItems {
		return n
	}
	return maxVisibleItems
}

func (m *Modal) MoveUp() {
	if len(m.Filtered) == 0 {
		return
	}
	if m.IsFocused {
		m.IsFocused = false
		m.Cursor = 0
		m.ScrollOffset = 0
		return
	}
	m.Cursor--
	if m.Cursor < 0 {
		m.Cursor = len(m.Filtered) - 1
	}
	m.clampScroll()
}

func (m *Modal) MoveDown() {
	if len(m.Filtered) == 0 {
		return
	}
	if m.IsFocused {
		m.IsFocused = false
		m.Cursor = 0
		m.ScrollOffset = 0
		return
	}
	m.Cursor++
	if m.Cursor >= len(m.Filtered) {
		m.Cursor = 0
	}
	m.clampScroll()
}

func (m *Modal) clampScroll() {
	n := m.VisibleCount()
	if m.Cursor < m.ScrollOffset {
		m.ScrollOffset = m.Cursor
	}
	if m.Cursor >= m.ScrollOffset+n {
		m.ScrollOffset = m.Cursor - n + 1
	}
}

func fold(s string) string {
	s = strings.ToLower(s)
	for _, r := range s {
		if r > 127 {
			s = strings.NewReplacer(
				"á", "a", "à", "a", "â", "a", "ã", "a", "ä", "a",
				"é", "e", "è", "e", "ê", "e", "ë", "e",
				"í", "i", "ì", "i", "î", "i", "ï", "i",
				"ó", "o", "ò", "o", "ô", "o", "õ", "o", "ö", "o",
				"ú", "u", "ù", "u", "û", "u", "ü", "u",
				"ç", "c", "ñ", "n",
			).Replace(s)
			break
		}
	}
	return s
}

func (m *Modal) Filter(query string) {
	m.FilterQuery = query
	if query == "" {
		m.Filtered = m.Items
	} else {
		q := fold(query)
		m.Filtered = nil
		for _, item := range m.Items {
			if strings.Contains(fold(item.Label), q) {
				m.Filtered = append(m.Filtered, item)
			}
		}
	}
	if len(m.Filtered) == 0 {
		m.Cursor = 0
	} else if m.Cursor >= len(m.Filtered) {
		m.Cursor = len(m.Filtered) - 1
	}
	m.ScrollOffset = 0
}

func (m *Modal) HandleRune(r rune) {
	if !m.IsFocused {
		return
	}
	m.Filter(m.FilterQuery + string(r))
}

func (m *Modal) HandleBackspace() {
	if !m.IsFocused {
		return
	}
	if len(m.FilterQuery) > 0 {
		runes := []rune(m.FilterQuery)
		m.Filter(string(runes[:len(runes)-1]))
	}
}

func (m *Modal) FocusInput() {
	m.IsFocused = true
}

func (m *Modal) FocusList() {
	if len(m.Filtered) > 0 {
		m.IsFocused = false
	}
}

func (m *Modal) Select() string {
	if m.Cursor >= 0 && m.Cursor < len(m.Filtered) {
		m.Selected = m.Filtered[m.Cursor].Key
	}
	m.Active = false
	return m.Selected
}

func (m *Modal) Close() {
	m.Active = false
}

func (m *Modal) SyncCursor(key string) {
	for i, item := range m.Filtered {
		if item.Key == key {
			m.Cursor = i
			m.clampScroll()
			return
		}
	}
	m.Cursor = 0
	m.ScrollOffset = 0
}

func (m *Modal) Render(bgColor, accentColor, fgColor string) string {
	if len(m.Items) == 0 {
		return ""
	}

	bg := lipgloss.Color(bgColor)
	accent := lipgloss.Color(accentColor)

	indicatorStyle := lipgloss.NewStyle().Foreground(accent).Background(bg)
	labelStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(fgColor)).Background(bg)
	inputStyle := lipgloss.NewStyle().Foreground(accent).Background(bg)
	cursorStyle := lipgloss.NewStyle().Foreground(bg).Background(accent)

	var buf strings.Builder

	title := " " + m.Title
	buf.WriteString(indicatorStyle.Render(title))
	buf.WriteString("\n")

	filterLabel := " / filter: "
	if m.IsFocused {
		display := m.FilterQuery
		if display == "" {
			display = " "
		}
		buf.WriteString(inputStyle.Render(filterLabel))
		buf.WriteString(inputStyle.Render(display))
		buf.WriteString(cursorStyle.Render(" "))
		buf.WriteString(inputStyle.Render(" "))
	} else {
		buf.WriteString(inputStyle.Render(filterLabel + m.FilterQuery + " "))
	}
	buf.WriteString("\n")

	if len(m.Filtered) == 0 {
		buf.WriteString(inputStyle.Render("   no matches"))
		buf.WriteString("\n")

		buf.WriteString("\n")
		buf.WriteString(indicatorStyle.Render(" / filter  ↑↓ nav  ↵ select  esc/q close"))
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

	if len(m.Filtered) > maxVisibleItems {
		buf.WriteString(indicatorStyle.Render(fmt.Sprintf("  (%d/%d)", m.ScrollOffset+1, len(m.Filtered))))
	}
	buf.WriteString("\n")

	start := m.ScrollOffset
	end := start + m.VisibleCount()
	if end > len(m.Filtered) {
		end = len(m.Filtered)
	}

	for _, item := range m.Filtered[start:end] {
		hasCursor := item.Key == m.Filtered[m.Cursor].Key
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
	buf.WriteString(indicatorStyle.Render(" / filter  ↑↓ nav  ↵ select  esc/q close"))

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

func (m *Modal) Overlay(base string, bgColor, borderColor, fgColor string) string {
	if !m.Active {
		return base
	}

	fg := m.Render(bgColor, borderColor, fgColor)
	return overlay.Composite(fg, base, overlay.Center, overlay.Center, 0, 0)
}
