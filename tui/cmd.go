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
		if err != nil {
			return diffMsg{nil, err}
		}

		untracked, err := git.Execute(context.Background(), runner, &git.LsFilesCommand{})
		if err != nil {
			return diffMsg{diffs, nil}
		}

		for _, fp := range untracked {
			d := git.UntrackedToSideBySideDiff(repoPath, fp)
			if d.NewPath != "" {
				diffs = append(diffs, d)
			}
		}

		return diffMsg{diffs, nil}
	}
}
