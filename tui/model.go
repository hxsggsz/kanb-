package tui

import (
	models "kanba/tui/models"
	tea "charm.land/bubbletea/v2"
	"kanba/git"
)

type screen int

const (
	screenDiff screen = iota
	screenHelp
)

const (
	statusBarHeight = 4
	lineNumColWidth = 4
)

const (
	sidebarMinWidth     = 15
	sidebarMaxWidth     = 35
	sidebarDefaultWidth = 25
	sidebarDenominator  = 4
	panelBorderWidth    = 0
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

	repoPath     string
	gitArgs      []string
	highlighter  *SyntaxHighlighter
	themeModal   *models.Modal
}

func (m *model) totalLines() int {
	return len(m.flatLines)
}

func (m *model) currentTheme() models.Theme {
	if m.themeModal != nil {
		return models.GetTheme(m.themeModal.Selected)
	}
	return models.GetTheme("rose-pine")
}

func New(repoPath string, gitArgs []string) tea.Model {
	themeItems := make([]models.ModalItem, 0, len(models.Themes))
	for _, k := range models.SortedThemeKeys() {
		t := models.Themes[k]
		themeItems = append(themeItems, models.ModalItem{Key: k, Label: t.Name})
	}
	themeModal := models.NewModal("Theme", themeItems)
	themeModal.Selected = "rose-pine"

	return &model{
		repoPath:    repoPath,
		gitArgs:     gitArgs,
		loading:     true,
		scroller:    NewScroller(),
		highlighter: NewSyntaxHighlighter(),
		themeModal:  themeModal,
	}
}

func (m *model) Init() tea.Cmd {
	return gitDiffCmd(m.repoPath, m.gitArgs)
}

func (m *model) visibleLines() int {
	return m.height - statusBarHeight
}
