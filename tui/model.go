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
	flatLines   []flatLine
	fileStats   []fileStat
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
	return len(m.flatLines)
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
	return m.height - statusBarHeight
}
