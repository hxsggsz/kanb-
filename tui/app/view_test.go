package app

import (
	"testing"

	"kanba/tui/diff"

	tea "charm.land/bubbletea/v2"
	"kanba/git"
)

func TestCursorAtEndOfFile(t *testing.T) {
	lines := make([]git.AlignedLine, 5)
	for i := range lines {
		lines[i] = git.AlignedLine{Kind: git.KindContext, OldLineNum: i + 1, NewLineNum: i + 1, OldContent: "line", NewContent: "line"}
	}
	h := git.AlignedHunk{
		Header: "@@ -1,5 +1,5 @@",
		Lines:  lines,
	}
	f := git.SideBySideDiff{
		NewPath: "test.txt",
		Status:  "M",
		Hunks:   []git.AlignedHunk{h},
	}

	m := &Model{
		diffs:     []git.SideBySideDiff{f},
		flatLines: diff.BuildFlatLines([]git.SideBySideDiff{f}),
		fileStats: diff.ComputeFileStats([]git.SideBySideDiff{f}),
		scroller:  diff.NewScroller(),
		width:     80,
		height:    24,
	}

	total := m.TotalLines()
	t.Logf("totalLines: %d", total)

	for i := 0; i < 100; i++ {
		m.handleDiffKeys(tea.KeyPressMsg{Code: tea.KeyDown})
	}

	if m.scroller.CursorLine() != total-1 {
		t.Fatalf("expected cursorLine=%d (last line), got %d", total-1, m.scroller.CursorLine())
	}

	m.renderContinuous(80, m.VisibleLines())
	if m.scroller.CursorLine() != total-1 {
		t.Fatalf("renderContinuous changed cursorLine: expected %d, got %d", total-1, m.scroller.CursorLine())
	}

	if m.scroller.CursorLine() >= m.TotalLines() {
		t.Fatalf("cursorLine %d >= totalLines %d after renderContinuous", m.scroller.CursorLine(), m.TotalLines())
	}
}

func TestViewCursorScrolling(t *testing.T) {
	lines := make([]git.AlignedLine, 6)
	for i := range lines {
		lines[i] = git.AlignedLine{Kind: git.KindContext, OldLineNum: i + 1, NewLineNum: i + 1, OldContent: "line", NewContent: "line"}
	}
	h := git.AlignedHunk{
		Header: "@@ -1,6 +1,6 @@",
		Lines:  lines,
	}
	f := git.SideBySideDiff{
		NewPath: "test.txt",
		Status:  "M",
		Hunks:   []git.AlignedHunk{h},
	}

	m := &Model{
		diffs:     []git.SideBySideDiff{f},
		flatLines: diff.BuildFlatLines([]git.SideBySideDiff{f}),
		fileStats: diff.ComputeFileStats([]git.SideBySideDiff{f}),
		scroller:  diff.NewScroller(),
		width:     80,
		height:    10,
	}

	total := m.TotalLines()
	vis := m.VisibleLines()

	for m.scroller.CursorLine() < total-1 {
		prevScroll := m.scroller.Scroll()
		m.handleDiffKeys(tea.KeyPressMsg{Code: tea.KeyDown})
		m.renderContinuous(80, m.VisibleLines())

		if m.scroller.Scroll() < 0 {
			t.Fatalf("negative scroll: %d", m.scroller.Scroll())
		}

		visPos := m.scroller.CursorLine() - m.scroller.Scroll()
		if visPos < 0 || visPos >= vis {
			t.Fatalf("DOWN: cursorLine=%d outside view [scroll=%d, scroll+vis=%d) visPos=%d",
				m.scroller.CursorLine(), prevScroll, m.scroller.Scroll()+vis, visPos)
		}
	}

	if m.scroller.CursorLine() != total-1 {
		t.Fatalf("expected cursor at last line %d, got %d", total-1, m.scroller.CursorLine())
	}

	for m.scroller.CursorLine() > 0 {
		prevScroll := m.scroller.Scroll()
		m.handleDiffKeys(tea.KeyPressMsg{Code: tea.KeyUp})
		m.renderContinuous(80, m.VisibleLines())

		if m.scroller.Scroll() < 0 {
			t.Fatalf("negative scroll: %d", m.scroller.Scroll())
		}

		visPos := m.scroller.CursorLine() - m.scroller.Scroll()
		if visPos < 0 || visPos >= vis {
			t.Fatalf("UP: cursorLine=%d outside view [scroll=%d, scroll+vis=%d) visPos=%d",
				m.scroller.CursorLine(), prevScroll, m.scroller.Scroll()+vis, visPos)
		}
	}

	if m.scroller.CursorLine() != 0 {
		t.Fatalf("expected cursor at line 0, got %d", m.scroller.CursorLine())
	}
}

func TestViewLayoutPreservesCursorWithinVisibleRange(t *testing.T) {
	lines := make([]git.AlignedLine, 8)
	for i := range lines {
		lines[i] = git.AlignedLine{Kind: git.KindContext, OldLineNum: i + 1, NewLineNum: i + 1, OldContent: "line", NewContent: "line"}
	}
	h := git.AlignedHunk{
		Header: "@@ -1,8 +1,8 @@",
		Lines:  lines,
	}
	f := git.SideBySideDiff{
		NewPath: "test.txt",
		Status:  "M",
		Hunks:   []git.AlignedHunk{h},
	}

	m := &Model{
		diffs:     []git.SideBySideDiff{f},
		flatLines: diff.BuildFlatLines([]git.SideBySideDiff{f}),
		fileStats: diff.ComputeFileStats([]git.SideBySideDiff{f}),
		scroller:  diff.NewScroller(),
		width:     80,
		height:    20,
	}

	total := m.TotalLines()
	vis := m.VisibleLines()

	if vis <= 0 {
		t.Fatal("vis must be > 0 for this test")
	}

	for m.scroller.CursorLine() < total-1 {
		m.handleDiffKeys(tea.KeyPressMsg{Code: tea.KeyDown})
		m.renderContinuous(80, m.VisibleLines())

		if m.scroller.CursorLine() < m.scroller.Scroll() || m.scroller.CursorLine() >= m.scroller.Scroll()+vis {
			t.Fatalf("cursorLine=%d outside [scroll=%d, scroll+vis=%d) after %d total presses",
				m.scroller.CursorLine(), m.scroller.Scroll(), m.scroller.Scroll()+vis, m.scroller.CursorLine()+1)
		}
	}

	if m.scroller.CursorLine() != total-1 {
		t.Fatalf("expected cursor at last line %d, got %d", total-1, m.scroller.CursorLine())
	}

	for m.scroller.CursorLine() > 0 {
		m.handleDiffKeys(tea.KeyPressMsg{Code: tea.KeyUp})
		m.renderContinuous(80, m.VisibleLines())

		if m.scroller.CursorLine() < m.scroller.Scroll() || m.scroller.CursorLine() >= m.scroller.Scroll()+vis {
			t.Fatalf("cursorLine=%d outside [scroll=%d, scroll+vis=%d)",
				m.scroller.CursorLine(), m.scroller.Scroll(), m.scroller.Scroll()+vis)
		}
	}

	if m.scroller.CursorLine() != 0 {
		t.Fatalf("expected cursor at first line, got %d", m.scroller.CursorLine())
	}
}
