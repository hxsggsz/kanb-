package app

import (
	"testing"

	"kanba/tui/diff"
)

func setupTestModel(lines []string, viewportWidth, viewportHeight int) *Model {
	m := &Model{
		width:       viewportWidth,
		height:      viewportHeight,
		highlighter: diff.NewSyntaxHighlighter(),
	}
	m.visibleLines = viewportHeight - 4
	m.flatLines = make([]diff.FlatLine, len(lines))
	for i := range lines {
		m.flatLines[i] = diff.FlatLine{}
	}
	m.scroller = diff.NewScroller()
	return m
}

func TestScrollForDifferentHeights(t *testing.T) {
	tests := []struct {
		name           string
		viewportWidth  int
		viewportHeight int
		scrollAmount   int
	}{
		{"width=80, height=24", 80, 24, 5},
		{"width=80, height=40", 80, 40, 10},
		{"width=120, height=60", 120, 60, 15},
		{"width=40, height=10", 40, 10, 3},
		{"height=5", 80, 5, 2},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			lines := make([]string, 100)
			for i := range lines {
				lines[i] = "line content"
			}
			m := setupTestModel(lines, tc.viewportWidth, tc.viewportHeight)

			for i := 0; i < tc.scrollAmount; i++ {
				m.scroller.MoveUp()
			}

			if m.scroller.Scroll() != 0 {
				t.Errorf("expected scroll=0 after scrolling up from 0, got %d", m.scroller.Scroll())
			}

			for i := 0; i < tc.scrollAmount; i++ {
				m.scroller.MoveDown(len(m.flatLines), m.visibleLines)
			}

			expected := tc.scrollAmount
			if m.scroller.Scroll() != expected {
				t.Errorf("expected scroll=%d, got %d", expected, m.scroller.Scroll())
			}
		})
	}
}

func TestScrollLimits(t *testing.T) {
	lines := make([]string, 30)
	for i := range lines {
		lines[i] = "content"
	}

	m := setupTestModel(lines, 80, 10)
	visibleLines := m.visibleLines
	totalLines := len(m.flatLines)

	for i := 0; i < 50; i++ {
		m.scroller.MoveDown(totalLines, visibleLines)
	}
	expectedScroll := totalLines - visibleLines
	if m.scroller.Scroll() != expectedScroll {
		t.Errorf("expected scroll to stop at %d, got %d", expectedScroll, m.scroller.Scroll())
	}

	for i := 0; i < 50; i++ {
		m.scroller.MoveUp()
	}
	if m.scroller.Scroll() != 0 {
		t.Errorf("expected scroll=0 after scrolling to top, got %d", m.scroller.Scroll())
	}
}
