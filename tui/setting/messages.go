package setting

import "kanba/git"

type DiffMsg struct {
	Diffs []git.SideBySideDiff
	Err   error
}
