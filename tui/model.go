package tui

import (
	tea "charm.land/bubbletea/v2"
	"kanba/git"
)

type screen int

const (
	screenDiff screen = iota
	screenHelp
)

const (
	statusBarHeight = 1
	borderHeight    = 2
	hunkHeaderLines = 2
	lineNumColWidth = 4
)

const (
	sidebarMinWidth     = 15
	sidebarMaxWidth     = 35
	sidebarDefaultWidth = 25
	sidebarDenominator  = 4
	panelBorderWidth    = 2
	panelMinWidth       = 10
)

type model struct {
	diffs   []git.FileDiff
	fileIdx int
	scroll  int
	screen  screen
	loading bool
	err     error
	width   int
	height  int

	repoPath string
	gitArgs  []string
}

func New(repoPath string, gitArgs []string) tea.Model {
	return &model{
		repoPath: repoPath,
		gitArgs:  gitArgs,
		loading:  true,
	}
}

func (m *model) Init() tea.Cmd {
	return gitDiffCmd(m.repoPath, m.gitArgs)
}

func (m *model) visibleLines() int {
	return m.height - (statusBarHeight + borderHeight)
}

func (m *model) maxScroll() int {
	if len(m.diffs) == 0 {
		return 0
	}
	f := m.diffs[m.fileIdx]
	totalLines := 0
	for _, h := range f.Hunks {
		totalLines += hunkHeaderLines + len(h.Lines)
	}
	max := totalLines - m.visibleLines()
	if max < 0 {
		return 0
	}
	return max
}
