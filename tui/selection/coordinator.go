package selection

import (
	"time"

	tea "charm.land/bubbletea/v2"
)

// Coordinator manages selection state and mouse event routing.
type Coordinator struct {
	state      State
	strategy   Strategy
	clickCount int
	lastClick  time.Time
	lastX      int
	lastY      int

	onCopy        func(string) tea.Cmd
	getLineContent func(line int) string
}

// NewCoordinator creates a new selection coordinator.
func NewCoordinator(onCopy func(string) tea.Cmd) *Coordinator {
	return &Coordinator{
		state:    IdleState{},
		strategy: CharacterStrategy{},
		onCopy:   onCopy,
	}
}

// SetLineContentProvider sets the callback used to retrieve line content
// for word boundary detection during double-click.
func (c *Coordinator) SetLineContentProvider(fn func(line int) string) {
	c.getLineContent = fn
}

// HandleClick processes a mouse click.
func (c *Coordinator) HandleClick(panel PanelSide, line, col int) tea.Cmd {
	now := time.Now()
	samePosition := abs(col-c.lastX) <= 2 && abs(line-c.lastY) <= 2

	if samePosition && now.Sub(c.lastClick) < 300*time.Millisecond {
		c.clickCount++
	} else {
		c.clickCount = 1
	}

	c.lastClick = now
	c.lastX = col
	c.lastY = line

	if c.clickCount >= 2 {
		c.clickCount = 0
		c.strategy = WordStrategy{}

		var boundary WordBoundary
		if c.getLineContent != nil {
			content := c.getLineContent(line)
			start, end := findWordBoundaries(content, col)
			boundary = WordBoundary{Start: start, End: end}
		} else {
			boundary = WordBoundary{Start: col, End: col}
		}

		c.state = c.state.HandleDoubleClick(c, panel, line, col, boundary)
		return nil
	}

	c.strategy = CharacterStrategy{}
	c.state = c.state.HandleClick(c, panel, line, col)
	return nil
}

// HandleDrag processes mouse drag.
func (c *Coordinator) HandleDrag(panel PanelSide, line, col int) {
	c.state = c.state.HandleDrag(c, panel, line, col)
}

// HandleRelease processes mouse release.
func (c *Coordinator) HandleRelease() tea.Cmd {
	var cmd tea.Cmd
	c.state, cmd = c.state.HandleRelease(c)
	return cmd
}

// Clear resets the selection.
func (c *Coordinator) Clear() {
	c.state = c.state.Clear()
	c.strategy = CharacterStrategy{}
}

// CurrentSelection returns the active selection (if any).
func (c *Coordinator) CurrentSelection() *Selection {
	switch st := c.state.(type) {
	case SelectingState:
		sel := st.Selection
		return &sel
	case SelectedState:
		sel := st.Selection
		return &sel
	default:
		return nil
	}
}

// HasSelection returns true if there's an active non-empty selection.
func (c *Coordinator) HasSelection() bool {
	sel := c.CurrentSelection()
	if sel == nil {
		return false
	}
	return !sel.Range.IsEmpty()
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}
