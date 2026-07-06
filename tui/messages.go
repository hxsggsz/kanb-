package tui

import "kanba/git"

type diffMsg struct {
	diffs []git.FileDiff
	err   error
}
