package app

import (
	models "kanba/tui/models"
	"kanba/tui/diff"
	"kanba/tui/setting"

	tea "charm.land/bubbletea/v2"
	"kanba/git"
)

const (
	statusBarHeight = 4
	panelBorderWidth = 0
	panelMinWidth   = 10
)

type screen int

const (
	screenDiff screen = iota
	screenHelp
)

type Model struct {
	diffs       []git.SideBySideDiff
	flatLines   []diff.FlatLine
	fileStats   []diff.FileStat
	scroller    *diff.Scroller
	screen      screen
	loading     bool
	err         error
	width       int
	height      int

	repoPath     string
	gitArgs      []string
	highlighter  *diff.SyntaxHighlighter
	themeModal   *models.Modal
}

func (m *Model) TotalLines() int {
	return len(m.flatLines)
}

func (m *Model) CurrentTheme() models.Theme {
	if m.themeModal != nil {
		return models.GetTheme(m.themeModal.Selected)
	}
	return models.GetTheme("rose-pine")
}

func New(repoPath string, gitArgs []string) *Model {
	themeItems := make([]models.ModalItem, 0, len(models.Themes))
	for _, k := range models.SortedThemeKeys() {
		t := models.Themes[k]
		themeItems = append(themeItems, models.ModalItem{Key: k, Label: t.Name})
	}
	themeModal := models.NewModal("Theme", themeItems)
	themeModal.Selected = "rose-pine"

	return &Model{
		repoPath:    repoPath,
		gitArgs:     gitArgs,
		loading:     true,
		scroller:    diff.NewScroller(),
		highlighter: diff.NewSyntaxHighlighter(),
		themeModal:  themeModal,
	}
}

func (m *Model) Init() tea.Cmd {
	return setting.GitDiffCmd(m.repoPath, m.gitArgs)
}

func (m *Model) VisibleLines() int {
	return m.height - statusBarHeight
}
