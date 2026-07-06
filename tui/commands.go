package tui

import (
	"net/http"
	"time"

	tea "charm.land/bubbletea/v2"
)

type statusMsg int
type errMsg struct{ err error }

func (e errMsg) Error() string { return e.err.Error() }

func checkServer(url string) tea.Cmd {
	return func() tea.Msg {
		c := &http.Client{Timeout: 10 * time.Second}
		res, err := c.Get(url)
		if err != nil {
			return errMsg{err}
		}
		return statusMsg(res.StatusCode)
	}
}

type tickMsg int

func doTick() tea.Cmd {
	return tea.Every(time.Second, func(t time.Time) tea.Msg {
		return tickMsg(t.Second())
	})
}
