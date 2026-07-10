package selection

import "testing"

func newTestCoordinator() *Coordinator {
	return &Coordinator{}
}
func TestIdleState_HandleClick(t *testing.T) {
	s := newTestCoordinator()
	state := IdleState{}

	next := state.HandleClick(s, PanelLeft, 3, 5)

	sel, ok := next.(SelectingState)
	if !ok {
		t.Fatalf("expected SelectingState, got %T", next)
	}
	if sel.Selection.Panel != PanelLeft {
		t.Errorf("Panel = %v, want PanelLeft", sel.Selection.Panel)
	}
	if sel.Selection.Range != (Range{StartLine: 3, StartCol: 5, EndLine: 3, EndCol: 5}) {
		t.Errorf("Range = %+v, want {StartLine:3 StartCol:5 EndLine:3 EndCol:5}", sel.Selection.Range)
	}
}

func TestIdleState_HandleDrag(t *testing.T) {
	s := newTestCoordinator()
	state := IdleState{}

	next := state.HandleDrag(s, PanelRight, 10, 20)

	if _, ok := next.(IdleState); !ok {
		t.Fatalf("expected IdleState, got %T", next)
	}
}

func TestIdleState_HandleRelease(t *testing.T) {
	s := newTestCoordinator()
	state := IdleState{}

	next, cmd := state.HandleRelease(s)

	if _, ok := next.(IdleState); !ok {
		t.Fatalf("expected IdleState, got %T", next)
	}
	if cmd != nil {
		t.Errorf("expected nil cmd, got %v", cmd)
	}
}

func TestIdleState_HandleDoubleClick(t *testing.T) {
	s := newTestCoordinator()
	state := IdleState{}
	boundary := WordBoundary{Start: 2, End: 8}

	next := state.HandleDoubleClick(s, PanelRight, 5, 4, boundary)

	sel, ok := next.(SelectedState)
	if !ok {
		t.Fatalf("expected SelectedState, got %T", next)
	}
	if sel.Selection.Panel != PanelRight {
		t.Errorf("Panel = %v, want PanelRight", sel.Selection.Panel)
	}
	if sel.Selection.Range != (Range{StartLine: 5, StartCol: 2, EndLine: 5, EndCol: 8}) {
		t.Errorf("Range = %+v, want {StartLine:5 StartCol:2 EndLine:5 EndCol:8}", sel.Selection.Range)
	}
}

func TestIdleState_Clear(t *testing.T) {
	state := IdleState{}
	next := state.Clear()
	if _, ok := next.(IdleState); !ok {
		t.Fatalf("expected IdleState, got %T", next)
	}
}

func TestSelectingState_HandleClick(t *testing.T) {
	s := newTestCoordinator()
	state := SelectingState{
		Selection: Selection{Panel: PanelLeft, Range: Range{StartLine: 0, StartCol: 0, EndLine: 5, EndCol: 10}},
	}

	next := state.HandleClick(s, PanelRight, 7, 3)

	sel, ok := next.(SelectingState)
	if !ok {
		t.Fatalf("expected SelectingState, got %T", next)
	}
	if sel.Selection.Panel != PanelRight {
		t.Errorf("Panel = %v, want PanelRight", sel.Selection.Panel)
	}
	if sel.Selection.Range != (Range{StartLine: 7, StartCol: 3, EndLine: 7, EndCol: 3}) {
		t.Errorf("Range = %+v, want {StartLine:7 StartCol:3 EndLine:7 EndCol:3}", sel.Selection.Range)
	}
}

func TestSelectingState_HandleDrag(t *testing.T) {
	s := newTestCoordinator()
	state := SelectingState{
		Selection: Selection{Panel: PanelLeft, Range: Range{StartLine: 0, StartCol: 0, EndLine: 0, EndCol: 0}},
	}

	next := state.HandleDrag(s, PanelLeft, 5, 10)

	sel, ok := next.(SelectingState)
	if !ok {
		t.Fatalf("expected SelectingState, got %T", next)
	}
	if sel.Selection.Range != (Range{StartLine: 0, StartCol: 0, EndLine: 5, EndCol: 10}) {
		t.Errorf("Range = %+v, want {StartLine:0 StartCol:0 EndLine:5 EndCol:10}", sel.Selection.Range)
	}
}

func TestSelectingState_HandleDrag_MultipleDrags(t *testing.T) {
	s := newTestCoordinator()
	state := SelectingState{
		Selection: Selection{Panel: PanelLeft, Range: Range{StartLine: 1, StartCol: 2, EndLine: 1, EndCol: 2}},
	}

	state = state.HandleDrag(s, PanelLeft, 3, 4).(SelectingState)
	state = state.HandleDrag(s, PanelLeft, 7, 8).(SelectingState)

	if state.Selection.Range != (Range{StartLine: 1, StartCol: 2, EndLine: 7, EndCol: 8}) {
		t.Errorf("Range after multiple drags = %+v, want {StartLine:1 StartCol:2 EndLine:7 EndCol:8}", state.Selection.Range)
	}
}

func TestSelectingState_HandleRelease(t *testing.T) {
	s := newTestCoordinator()
	state := SelectingState{
		Selection: Selection{Panel: PanelLeft, Range: Range{StartLine: 0, StartCol: 0, EndLine: 5, EndCol: 10}},
	}

	next, cmd := state.HandleRelease(s)

	sel, ok := next.(SelectedState)
	if !ok {
		t.Fatalf("expected SelectedState, got %T", next)
	}
	if sel.Selection.Range != (Range{StartLine: 0, StartCol: 0, EndLine: 5, EndCol: 10}) {
		t.Errorf("Range = %+v, want {StartLine:0 StartCol:0 EndLine:5 EndCol:10}", sel.Selection.Range)
	}
	if cmd != nil {
		t.Errorf("expected nil cmd, got %v", cmd)
	}
}

func TestSelectingState_HandleDoubleClick(t *testing.T) {
	s := newTestCoordinator()
	state := SelectingState{
		Selection: Selection{Panel: PanelLeft, Range: Range{StartLine: 0, StartCol: 0, EndLine: 2, EndCol: 5}},
	}
	boundary := WordBoundary{Start: 10, End: 20}

	next := state.HandleDoubleClick(s, PanelRight, 4, 15, boundary)

	sel, ok := next.(SelectedState)
	if !ok {
		t.Fatalf("expected SelectedState, got %T", next)
	}
	if sel.Selection.Panel != PanelRight {
		t.Errorf("Panel = %v, want PanelRight", sel.Selection.Panel)
	}
	if sel.Selection.Range != (Range{StartLine: 4, StartCol: 10, EndLine: 4, EndCol: 20}) {
		t.Errorf("Range = %+v, want {StartLine:4 StartCol:10 EndLine:4 EndCol:20}", sel.Selection.Range)
	}
}

func TestSelectingState_Clear(t *testing.T) {
	state := SelectingState{
		Selection: Selection{Panel: PanelLeft, Range: Range{StartLine: 0, StartCol: 0, EndLine: 5, EndCol: 10}},
	}

	next := state.Clear()

	if _, ok := next.(IdleState); !ok {
		t.Fatalf("expected IdleState, got %T", next)
	}
}

func TestSelectedState_HandleClick(t *testing.T) {
	s := newTestCoordinator()
	state := SelectedState{
		Selection: Selection{Panel: PanelLeft, Range: Range{StartLine: 0, StartCol: 0, EndLine: 5, EndCol: 10}},
	}

	next := state.HandleClick(s, PanelRight, 8, 2)

	sel, ok := next.(SelectingState)
	if !ok {
		t.Fatalf("expected SelectingState, got %T", next)
	}
	if sel.Selection.Panel != PanelRight {
		t.Errorf("Panel = %v, want PanelRight", sel.Selection.Panel)
	}
	if sel.Selection.Range != (Range{StartLine: 8, StartCol: 2, EndLine: 8, EndCol: 2}) {
		t.Errorf("Range = %+v, want {StartLine:8 StartCol:2 EndLine:8 EndCol:2}", sel.Selection.Range)
	}
}

func TestSelectedState_HandleDrag(t *testing.T) {
	coordinator := newTestCoordinator()
	state := SelectedState{
		Selection: Selection{Panel: PanelLeft, Range: Range{StartLine: 0, StartCol: 0, EndLine: 5, EndCol: 10}},
	}

	next := state.HandleDrag(coordinator, PanelLeft, 99, 99)

	sel, ok := next.(SelectedState)
	if !ok {
		t.Fatalf("expected SelectedState, got %T", next)
	}
	// Selection should be unchanged
	if sel.Selection.Range != (Range{StartLine: 0, StartCol: 0, EndLine: 5, EndCol: 10}) {
		t.Errorf("Range = %+v, want unchanged {StartLine:0 StartCol:0 EndLine:5 EndCol:10}", sel.Selection.Range)
	}
}

func TestSelectedState_HandleRelease(t *testing.T) {
	s := newTestCoordinator()
	state := SelectedState{
		Selection: Selection{Panel: PanelLeft, Range: Range{StartLine: 0, StartCol: 0, EndLine: 5, EndCol: 10}},
	}

	next, cmd := state.HandleRelease(s)

	sel, ok := next.(SelectedState)
	if !ok {
		t.Fatalf("expected SelectedState, got %T", next)
	}
	if sel.Selection.Range != (Range{StartLine: 0, StartCol: 0, EndLine: 5, EndCol: 10}) {
		t.Errorf("Range = %+v, want unchanged", sel.Selection.Range)
	}
	if cmd != nil {
		t.Errorf("expected nil cmd, got %v", cmd)
	}
}

func TestSelectedState_HandleDoubleClick(t *testing.T) {
	s := newTestCoordinator()
	state := SelectedState{
		Selection: Selection{Panel: PanelLeft, Range: Range{StartLine: 0, StartCol: 0, EndLine: 5, EndCol: 10}},
	}
	boundary := WordBoundary{Start: 3, End: 9}

	next := state.HandleDoubleClick(s, PanelLeft, 2, 5, boundary)

	sel, ok := next.(SelectedState)
	if !ok {
		t.Fatalf("expected SelectedState, got %T", next)
	}
	if sel.Selection.Range != (Range{StartLine: 2, StartCol: 3, EndLine: 2, EndCol: 9}) {
		t.Errorf("Range = %+v, want {StartLine:2 StartCol:3 EndLine:2 EndCol:9}", sel.Selection.Range)
	}
}

func TestSelectedState_Clear(t *testing.T) {
	state := SelectedState{
		Selection: Selection{Panel: PanelLeft, Range: Range{StartLine: 0, StartCol: 0, EndLine: 5, EndCol: 10}},
	}

	next := state.Clear()

	if _, ok := next.(IdleState); !ok {
		t.Fatalf("expected IdleState, got %T", next)
	}
}

func TestFullLifecycle(t *testing.T) {
	s := newTestCoordinator()
	var state State = IdleState{}

	// Click to start selection
	state = state.HandleClick(s, PanelLeft, 1, 2)
	if _, ok := state.(SelectingState); !ok {
		t.Fatalf("after click: expected SelectingState, got %T", state)
	}

	// Drag to extend
	state = state.HandleDrag(s, PanelLeft, 3, 4)
	sel, ok := state.(SelectingState)
	if !ok {
		t.Fatalf("after drag: expected SelectingState, got %T", state)
	}
	if sel.Selection.Range.EndLine != 3 || sel.Selection.Range.EndCol != 4 {
		t.Fatalf("after drag: EndLine=%d EndCol=%d, want 3,4", sel.Selection.Range.EndLine, sel.Selection.Range.EndCol)
	}

	// Release to finalize
	state, cmd := state.(SelectingState).HandleRelease(s)
	if cmd != nil {
		t.Fatalf("after release: expected nil cmd, got %v", cmd)
	}
	if _, ok := state.(SelectedState); !ok {
		t.Fatalf("after release: expected SelectedState, got %T", state)
	}

	// Click again to start new selection
	state = state.HandleClick(s, PanelRight, 10, 20)
	if _, ok := state.(SelectingState); !ok {
		t.Fatalf("after second click: expected SelectingState, got %T", state)
	}

	// Clear
	state = state.Clear()
	if _, ok := state.(IdleState); !ok {
		t.Fatalf("after clear: expected IdleState, got %T", state)
	}
}
