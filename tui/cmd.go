package tui

import (
	"context"

	tea "charm.land/bubbletea/v2"
	"kanba/git"
)

func gitDiffCmd(repoPath string, args []string) tea.Cmd {
	return func() tea.Msg {
		runner := git.NewGitRunner(repoPath)
		cmd := &git.DiffCommand{
			DiffArgs: git.DiffArgs{Show: false, Args: args},
			Parser:   git.NewUnifiedParser(),
			Aligner:  &git.UnifiedAligner{},
		}
		diffs, err := git.Execute(context.Background(), runner, cmd)
		return diffMsg{diffs, err}
	}
}
