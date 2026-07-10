package selection

import (
	"strings"
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
		return c.copyIfSelected()
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
	c.state, _ = c.state.HandleRelease(c)
	return c.copyIfSelected()
}

// copyIfSelected extracts selected text and returns a DelayedCopyCmd if valid.
func (c *Coordinator) copyIfSelected() tea.Cmd {
	sel := c.CurrentSelection()
	if sel == nil || sel.Range.IsEmpty() {
		return nil
	}

	if c.getLineContent == nil {
		return nil
	}

	text := c.extractSelectedText(sel)
	if text == "" {
		return nil
	}

	return DelayedCopyCmd(text)
}

// extractSelectedText extracts the plain text from the current selection.
func (c *Coordinator) extractSelectedText(sel *Selection) string {
	normalized := sel.Range.Normalized()
	var lines []string
	for lineIdx := normalized.StartLine; lineIdx <= normalized.EndLine; lineIdx++ {
		line := c.getLineContent(lineIdx)
		startCol := 0
		endCol := len([]rune(line))
		if lineIdx == normalized.StartLine {
			startCol = normalized.StartCol
		}
		if lineIdx == normalized.EndLine {
			endCol = normalized.EndCol
		}
		if startCol < endCol && startCol < len([]rune(line)) {
			runes := []rune(line)
			if endCol > len(runes) {
				endCol = len(runes)
			}
			lines = append(lines, string(runes[startCol:endCol]))
		}
	}
	return strings.Join(lines, "\n")
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
