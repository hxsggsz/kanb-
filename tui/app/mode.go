package app

import (
	tea "charm.land/bubbletea/v2"
)

type ModeType int

const (
	ModeDefault      ModeType = iota
	ModeDiffOnly
	ModeRightPanel
)

const (
	modeBreakpointDefault     = 160
	modeBreakpointDiffOnly    = 100
)

type ViewMode interface {
	Render(m *Model) string
	HandleInput(m *Model, msg tea.KeyPressMsg) (tea.Model, tea.Cmd)
	Type() ModeType
}

type ModeFactory struct{}

func (f *ModeFactory) FromWidth(width int) ViewMode {
	switch {
	case width >= modeBreakpointDefault:
		return &DefaultMode{}
	case width >= modeBreakpointDiffOnly:
		return &DiffOnlyMode{}
	default:
		return &RightPanelMode{}
	}
}
