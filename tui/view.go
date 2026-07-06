package tui

import (
	"fmt"

	tea "charm.land/bubbletea/v2"
)

func (m *model) View() tea.View {
	v := tea.NewView("")

	if m.err != nil {
		v.SetContent(errorStyle.Render(fmt.Sprintf("Error: %v", m.err)))
		return v
	}

	v.AltScreen = true
	v.MouseMode = tea.MouseModeCellMotion
	v.WindowTitle = "kanba"

	switch m.screen {
	case screenDiff:
		v.SetContent(m.diffView())
	case screenHelp:
		v.SetContent(m.helpView())
	}

	return v
}

func (m *model) diffView() string {
	if m.loading {
		return appStyle.Render(spinnerStyle.Render("Loading..."))
	}
	return appStyle.Render(m.diffsView())
}

func (m *model) diffsView() string {
	if len(m.diffs) == 0 {
		return "No changes"
	}
	return titleStyle.Render("kanba")
}

func (m *model) helpView() string {
	content := titleStyle.Render("Help") + "\n\n"

	bindings := []struct {
		key  string
		desc string
	}{
		{"↑/k", "Scroll up"},
		{"↓/j", "Scroll down"},
		{"n/p", "Next/prev file"},
		{"g/G", "Top/bottom"},
		{"?", "Toggle help"},
		{"q/ctrl+c", "Quit"},
	}

	for _, b := range bindings {
		content += fmt.Sprintf("%-12s %s\n", itemStyle.Render(b.key), b.desc)
	}

	content += "\n" + helpStyle.Render("esc or ? to close")
	return appStyle.Render(content)
}
