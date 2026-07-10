package selection

import tea "charm.land/bubbletea/v2"

type State interface {
	HandleClick(s *Coordinator, panel PanelSide, line, col int) State
	HandleDrag(s *Coordinator, panel PanelSide, line, col int) State
	HandleRelease(s *Coordinator) (State, tea.Cmd)
	HandleDoubleClick(s *Coordinator, panel PanelSide, line, col int, boundaries WordBoundary) State
	Clear() State
}

type WordBoundary struct {
	Start int
	End   int
}

// IdleState - no selection active
type IdleState struct{}

func (IdleState) HandleClick(s *Coordinator, panel PanelSide, line, col int) State {
	return SelectingState{
		Selection: Selection{
			Panel: panel,
			Range: Range{
				StartLine: line,
				StartCol:  col,
				EndLine:   line,
				EndCol:    col,
			},
		},
	}
}

func (IdleState) HandleDrag(s *Coordinator, panel PanelSide, line, col int) State {
	return IdleState{}
}

func (IdleState) HandleRelease(s *Coordinator) (State, tea.Cmd) {
	return IdleState{}, nil
}

func (IdleState) HandleDoubleClick(s *Coordinator, panel PanelSide, line, col int, boundaries WordBoundary) State {
	return SelectedState{
		Selection: Selection{
			Panel: panel,
			Range: Range{
				StartLine: line,
				StartCol:  boundaries.Start,
				EndLine:   line,
				EndCol:    boundaries.End,
			},
		},
	}
}

func (IdleState) Clear() State {
	return IdleState{}
}

// SelectingState - mouse down, dragging
type SelectingState struct {
	Selection Selection
}

func (SelectingState) HandleClick(s *Coordinator, panel PanelSide, line, col int) State {
	return SelectingState{
		Selection: Selection{
			Panel: panel,
			Range: Range{
				StartLine: line,
				StartCol:  col,
				EndLine:   line,
				EndCol:    col,
			},
		},
	}
}

func (st SelectingState) HandleDrag(s *Coordinator, panel PanelSide, line, col int) State {
	st.Selection.Range.EndLine = line
	st.Selection.Range.EndCol = col
	return st
}

func (st SelectingState) HandleRelease(s *Coordinator) (State, tea.Cmd) {
	return SelectedState{Selection: st.Selection}, nil
}

func (st SelectingState) HandleDoubleClick(s *Coordinator, panel PanelSide, line, col int, boundaries WordBoundary) State {
	return SelectedState{
		Selection: Selection{
			Panel: panel,
			Range: Range{
				StartLine: line,
				StartCol:  boundaries.Start,
				EndLine:   line,
				EndCol:    boundaries.End,
			},
		},
	}
}

func (st SelectingState) Clear() State {
	return IdleState{}
}

// SelectedState - selection complete, awaiting copy or new click
type SelectedState struct {
	Selection Selection
}

func (SelectedState) HandleClick(s *Coordinator, panel PanelSide, line, col int) State {
	return SelectingState{
		Selection: Selection{
			Panel: panel,
			Range: Range{
				StartLine: line,
				StartCol:  col,
				EndLine:   line,
				EndCol:    col,
			},
		},
	}
}

func (st SelectedState) HandleDrag(s *Coordinator, panel PanelSide, line, col int) State {
	st.Selection.Range.EndLine = line
	st.Selection.Range.EndCol = col
	return st
}

func (st SelectedState) HandleRelease(s *Coordinator) (State, tea.Cmd) {
	return st, nil
}

func (st SelectedState) HandleDoubleClick(s *Coordinator, panel PanelSide, line, col int, boundaries WordBoundary) State {
	return SelectedState{
		Selection: Selection{
			Panel: panel,
			Range: Range{
				StartLine: line,
				StartCol:  boundaries.Start,
				EndLine:   line,
				EndCol:    boundaries.End,
			},
		},
	}
}

func (st SelectedState) Clear() State {
	return IdleState{}
}
