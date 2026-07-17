package setting

import "kanba/git"

type DiffMsg struct {
	Diffs []git.SideBySideDiff
	Err   error
}

type UpdateCheckMsg struct {
	Version   string
	Available bool
}

type UpdateInstallMsg struct {
	Err error
}
