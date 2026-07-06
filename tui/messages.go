package tui

import "kanba/git"

type diffMsg struct {
	diffs []git.SideBySideDiff
	err   error
}
