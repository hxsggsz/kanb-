package tui

import (
	tea "charm.land/bubbletea/v2"
	"kanba/git"
)

func gitDiffCmd(repoPath string, args []string) tea.Cmd {
	return func() tea.Msg {
		diffs, err := git.Diff(repoPath, args)
		return diffMsg{diffs, err}
	}
}
