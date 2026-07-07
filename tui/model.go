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
	diffs       []git.SideBySideDiff
	fileIdx     int
	scroller    *Scroller
	screen      screen
	loading     bool
	err         error
	width       int
	height      int

	repoPath    string
	gitArgs     []string
	highlighter *SyntaxHighlighter
}

func (m *model) totalLines() int {
	if len(m.diffs) == 0 {
		return 0
	}
	f := m.diffs[m.fileIdx]
	total := 0
	for _, h := range f.Hunks {
		total += len(h.Lines)
	}
	return total
}

func New(repoPath string, gitArgs []string) tea.Model {
	return &model{
		repoPath:    repoPath,
		gitArgs:     gitArgs,
		loading:     true,
		scroller:    NewScroller(),
		highlighter: NewSyntaxHighlighter(),
	}
}

func (m *model) Init() tea.Cmd {
	return gitDiffCmd(m.repoPath, m.gitArgs)
}

func (m *model) visibleLines() int {
	return m.height - (statusBarHeight + borderHeight)
}
