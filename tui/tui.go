package tui

import (
	tea "charm.land/bubbletea/v2"
	"kanba/tui/app"
)

func New(repoPath string, gitArgs []string) tea.Model {
	return app.New(repoPath, gitArgs)
}
