package selection

import (
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/atotto/clipboard"
)

const copyDelay = 400 * time.Millisecond

type CopyMsg struct{}

func DelayedCopyCmd() tea.Cmd {
	return tea.Tick(copyDelay, func(t time.Time) tea.Msg {
		return CopyMsg{}
	})
}

func CopyToClipboard(content string) error {
	return clipboard.WriteAll(content)
}
