package selection

import (
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/atotto/clipboard"
)

const copyDelay = 400 * time.Millisecond

// CopyMsg is sent when text should be copied
type CopyMsg struct {
	Content string
}

// DelayedCopyCmd returns a command that schedules a copy after the threshold
func DelayedCopyCmd(content string) tea.Cmd {
	return tea.Tick(copyDelay, func(t time.Time) tea.Msg {
		return CopyMsg{Content: content}
	})
}

// CopyToClipboard writes text to the system clipboard
func CopyToClipboard(content string) error {
	return clipboard.WriteAll(content)
}
