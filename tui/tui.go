package tui

import (
	tea "charm.land/bubbletea/v2"
	"kanba/config"
	"kanba/tui/app"
)

func New(repoPath string, gitArgs []string, cfg *config.Config) tea.Model {
	return app.New(repoPath, gitArgs, cfg.Theme)
}
