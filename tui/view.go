package tui

import (
	"fmt"

	tea "charm.land/bubbletea/v2"
)

func (m model) View() tea.View {
	v := tea.NewView("")

	if m.err != nil {
		v.SetContent(errorStyle.Render(fmt.Sprintf("Error: %v", m.err)))
		return v
	}

	v.AltScreen = true
	v.MouseMode = tea.MouseModeCellMotion
	v.WindowTitle = "kanba"

	switch m.currentScreen {
	case screenHome:
		v.SetContent(m.homeView())
	case screenDetail:
		v.SetContent(m.detailView())
	case screenHelp:
		v.SetContent(m.helpView())
	}

	return v
}

func (m model) homeView() string {
	s := titleStyle.Render("kanba")
	s += "\n\n"

	for i, item := range m.items {
		cursor := "  "
		if m.cursor == i {
			cursor = "▸ "
			s += selectedItemStyle.Render(fmt.Sprintf("%s%s", cursor, item.title)) + "\n"
		} else {
			s += itemStyle.Render(fmt.Sprintf("%s%s", cursor, item.title)) + "\n"
		}
	}

	if m.loading {
		s += "\n" + spinnerStyle.Render("Loading...") + "\n"
	}

	s += "\n" + helpStyle.Render("↑/k ↓/j • enter • ? help • q/ctrl+c quit")
	return appStyle.Render(s)
}

func (m model) detailView() string {
	if m.cursor < 0 || m.cursor >= len(m.items) {
		return ""
	}

	item := m.items[m.cursor]
	content := titleStyle.Render(item.title) + "\n\n"
	content += detailStyle.Render(item.description)
	content += "\n\n" + helpStyle.Render("esc to go back")

	return appStyle.Render(content)
}

func (m model) helpView() string {
	content := titleStyle.Render("Help") + "\n\n"

	bindings := []struct {
		key  string
		desc string
	}{
		{"↑/k", "Move cursor up"},
		{"↓/j", "Move cursor down"},
		{"enter", "Select item"},
		{"esc", "Go back"},
		{"?", "Toggle help"},
		{"r", "Refresh"},
		{"q/ctrl+c", "Quit"},
	}

	for _, b := range bindings {
		content += fmt.Sprintf("%-12s %s\n", itemStyle.Render(b.key), b.desc)
	}

	content += "\n" + helpStyle.Render("esc or ? to close")
	return appStyle.Render(content)
}
