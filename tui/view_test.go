package tui

import (
	"testing"

	tea "charm.land/bubbletea/v2"
	"kanba/git"
)

func TestCursorAtEndOfFile(t *testing.T) {
	h := git.AlignedHunk{
		Header: "@@ -1,5 +1,6 @@",
		Lines: []git.AlignedLine{
			{Kind: git.KindContext, OldLineNum: 1, NewLineNum: 1, OldContent: "line1", NewContent: "line1"},
			{Kind: git.KindContext, OldLineNum: 2, NewLineNum: 2, OldContent: "line2", NewContent: "line2"},
			{Kind: git.KindContext, OldLineNum: 3, NewLineNum: 3, OldContent: "line3", NewContent: "line3"},
			{Kind: git.KindContext, OldLineNum: 4, NewLineNum: 4, OldContent: "line4", NewContent: "line4"},
			{Kind: git.KindAdded, OldLineNum: 0, NewLineNum: 5, NewContent: "line5new"},
		},
	}
	f := git.SideBySideDiff{
		NewPath: "test.txt",
		Status:  "M",
		Hunks:   []git.AlignedHunk{h},
	}

	m := &model{
		diffs:    []git.SideBySideDiff{f},
		fileIdx:  0,
		scroller: NewScroller(),
		width:    80,
		height:   24,
	}

	total := m.totalLines()
	t.Logf("totalLines: %d", total)

	nav := func(steps int) {
		for i := 0; i < steps; i++ {
			m.handleDiffKeys(tea.KeyPressMsg{Code: tea.KeyDown})
		}
	}

	nav(5)
	if m.scroller.CursorLine() != 4 {
		t.Fatalf("expected cursorLine=4, got %d", m.scroller.CursorLine())
	}

	nav(1)
	if m.scroller.CursorLine() != 4 {
		t.Fatalf("cursor moved past end: expected 4, got %d", m.scroller.CursorLine())
	}

	nav(100)
	if m.scroller.CursorLine() != 4 {
		t.Fatalf("cursor moved past end after 100 presses: expected 4, got %d", m.scroller.CursorLine())
	}

	m.renderFile(f, 80, m.visibleLines())
	if m.scroller.CursorLine() != 4 {
		t.Fatalf("renderFile changed cursorLine: expected 4, got %d", m.scroller.CursorLine())
	}
	if m.scroller.CursorLine() >= m.totalLines() {
		t.Fatalf("cursorLine %d >= totalLines %d after renderFile", m.scroller.CursorLine(), m.totalLines())
	}
}

func TestScrollAdvancesWithCursor(t *testing.T) {
	lines := make([]git.AlignedLine, 200)
	for i := range lines {
		lines[i] = git.AlignedLine{Kind: git.KindContext, OldLineNum: i + 1, NewLineNum: i + 1, OldContent: "line", NewContent: "line"}
	}
	h := git.AlignedHunk{
		Header: "@@ -1,200 +1,200 @@",
		Lines:  lines,
	}
	f := git.SideBySideDiff{
		NewPath: "test.txt",
		Status:  "M",
		Hunks:   []git.AlignedHunk{h},
	}

	m := &model{
		diffs:    []git.SideBySideDiff{f},
		fileIdx:  0,
		scroller: NewScroller(),
		width:    80,
		height:   24,
	}

	total := m.totalLines()
	t.Logf("totalLines: %d, expected: %d", total, 200)

	vis := m.visibleLines()

	for i := 0; i < total-1; i++ {
		m.handleDiffKeys(tea.KeyPressMsg{Code: tea.KeyDown})

		m.renderFile(f, 80, m.visibleLines())

		visPos := m.scroller.CursorLine() - m.scroller.Scroll()
		if visPos < 0 || visPos >= vis {
			t.Fatalf("step %d: cursor visible position %d (cursorLine=%d, scroll=%d) out of view [0, %d)",
				i+1, visPos, m.scroller.CursorLine(), m.scroller.Scroll(), vis)
		}

		if m.scroller.CursorLine() >= vis && m.scroller.Scroll() == 0 {
			t.Fatalf("step %d: scroll is 0 but cursorLine=%d is past visible area (vis=%d)",
				i+1, m.scroller.CursorLine(), vis)
		}
	}

	t.Logf("final state: cursorLine=%d, scroll=%d, total=%d, vis=%d",
		m.scroller.CursorLine(), m.scroller.Scroll(), total, vis)
	if m.scroller.CursorLine() != total-1 {
		t.Fatalf("expected cursorLine=%d, got %d", total-1, m.scroller.CursorLine())
	}
}

func TestScrollDoesNotGetStuck(t *testing.T) {
	lines := make([]git.AlignedLine, 100)
	for i := range lines {
		lines[i] = git.AlignedLine{Kind: git.KindContext, OldLineNum: i + 1, NewLineNum: i + 1, OldContent: "line", NewContent: "line"}
	}
	h := git.AlignedHunk{
		Header: "@@ -1,100 +1,100 @@",
		Lines:  lines,
	}
	f := git.SideBySideDiff{
		NewPath: "test.txt",
		Status:  "M",
		Hunks:   []git.AlignedHunk{h},
	}

	m := &model{
		diffs:    []git.SideBySideDiff{f},
		fileIdx:  0,
		scroller: NewScroller(),
		width:    80,
		height:   24,
	}

	total := m.totalLines()
	vis := m.visibleLines()
	maxScroll := total - vis

	t.Logf("total=%d, vis=%d, maxScroll=%d", total, vis, maxScroll)

	var lastScroll int = -1
	scrollStallCount := 0

	for i := 0; i < total; i++ {
		if i > 0 {
			m.handleDiffKeys(tea.KeyPressMsg{Code: tea.KeyDown})
		}

		m.renderFile(f, 80, m.visibleLines())

		if m.scroller.CursorLine() >= vis {
			if m.scroller.Scroll() == lastScroll {
				scrollStallCount++
			} else {
				scrollStallCount = 0
			}

			if scrollStallCount > 5 && m.scroller.Scroll() < maxScroll {
				t.Fatalf("step %d: scroll stuck at %d for %d steps, maxScroll=%d, cursorLine=%d",
					i, m.scroller.Scroll(), scrollStallCount, maxScroll, m.scroller.CursorLine())
			}
		}

		lastScroll = m.scroller.Scroll()
	}

	t.Logf("final: cursorLine=%d, scroll=%d, maxScroll=%d",
		m.scroller.CursorLine(), m.scroller.Scroll(), maxScroll)
}

func TestScrollKeepsCursorInViewForLargeFile(t *testing.T) {
	lines := make([]git.AlignedLine, 500)
	for i := range lines {
		lines[i] = git.AlignedLine{Kind: git.KindContext, OldLineNum: i + 1, NewLineNum: i + 1, OldContent: "line", NewContent: "line"}
	}
	h := git.AlignedHunk{
		Header: "@@ -1,500 +1,500 @@",
		Lines:  lines,
	}
	f := git.SideBySideDiff{
		NewPath: "test.txt",
		Status:  "M",
		Hunks:   []git.AlignedHunk{h},
	}

	m := &model{
		diffs:    []git.SideBySideDiff{f},
		fileIdx:  0,
		scroller: NewScroller(),
		width:    80,
		height:   24,
	}

	total := m.totalLines()
	vis := m.visibleLines()

	for i := 0; i < total-1; i++ {
		m.handleDiffKeys(tea.KeyPressMsg{Code: tea.KeyDown})
		m.renderFile(f, 80, m.visibleLines())

		if m.scroller.CursorLine() < m.scroller.Scroll() || m.scroller.CursorLine() >= m.scroller.Scroll()+vis {
			t.Fatalf("cursorLine=%d outside view [scroll=%d, scroll+vis=%d) at step %d",
				m.scroller.CursorLine(), m.scroller.Scroll(), m.scroller.Scroll()+vis, i+1)
		}
	}

	if m.scroller.CursorLine() != total-1 {
		t.Fatalf("expected cursor at last line %d, got %d", total-1, m.scroller.CursorLine())
	}
}

func TestScrollUpFromBottom(t *testing.T) {
	lines := make([]git.AlignedLine, 100)
	for i := range lines {
		lines[i] = git.AlignedLine{Kind: git.KindContext, OldLineNum: i + 1, NewLineNum: i + 1, OldContent: "line", NewContent: "line"}
	}
	h := git.AlignedHunk{
		Header: "@@ -1,100 +1,100 @@",
		Lines:  lines,
	}
	f := git.SideBySideDiff{
		NewPath: "test.txt",
		Status:  "M",
		Hunks:   []git.AlignedHunk{h},
	}

	m := &model{
		diffs:    []git.SideBySideDiff{f},
		fileIdx:  0,
		scroller: NewScroller(),
		width:    80,
		height:   24,
	}

	total := m.totalLines()

	for i := 0; i < total-1; i++ {
		m.handleDiffKeys(tea.KeyPressMsg{Code: tea.KeyDown})
	}
	m.renderFile(f, 80, m.visibleLines())

	t.Logf("at bottom: cursorLine=%d, scroll=%d", m.scroller.CursorLine(), m.scroller.Scroll())

	for i := 0; i < total-1; i++ {
		m.handleDiffKeys(tea.KeyPressMsg{Code: tea.KeyUp})
		m.renderFile(f, 80, m.visibleLines())

		if m.scroller.Scroll() < 0 {
			t.Fatalf("negative scroll at step %d: %d", i+1, m.scroller.Scroll())
		}
	}

	if m.scroller.CursorLine() != 0 {
		t.Fatalf("expected cursor at line 0, got %d", m.scroller.CursorLine())
	}
	if m.scroller.Scroll() != 0 {
		t.Fatalf("expected scroll=0 at top, got %d", m.scroller.Scroll())
	}
}

func TestScrollHysteresisCursorVisiblePosition(t *testing.T) {
	lines := make([]git.AlignedLine, 50)
	for i := range lines {
		lines[i] = git.AlignedLine{Kind: git.KindContext, OldLineNum: i + 1, NewLineNum: i + 1, OldContent: "line", NewContent: "line"}
	}
	h := git.AlignedHunk{
		Header: "@@ -1,50 +1,50 @@",
		Lines:  lines,
	}
	f := git.SideBySideDiff{
		NewPath: "test.txt",
		Status:  "M",
		Hunks:   []git.AlignedHunk{h},
	}

	m := &model{
		diffs:    []git.SideBySideDiff{f},
		fileIdx:  0,
		scroller: NewScroller(),
		width:    80,
		height:   24,
	}

	total := m.totalLines()

	for i := 0; i < total-1; i++ {
		m.handleDiffKeys(tea.KeyPressMsg{Code: tea.KeyDown})
		m.renderFile(f, 80, m.visibleLines())
		visPos := m.scroller.CursorLine() - m.scroller.Scroll()
		if visPos < 0 || visPos >= m.visibleLines() {
			t.Fatalf("DOWN step %d: cursor visible position %d out of bounds (cursor=%d, scroll=%d, vis=%d)",
				i+1, visPos, m.scroller.CursorLine(), m.scroller.Scroll(), m.visibleLines())
		}
	}

	for i := 0; i < total-1; i++ {
		m.handleDiffKeys(tea.KeyPressMsg{Code: tea.KeyUp})
		m.renderFile(f, 80, m.visibleLines())
		visPos := m.scroller.CursorLine() - m.scroller.Scroll()
		if visPos < 0 || visPos >= m.visibleLines() {
			t.Fatalf("UP step %d: cursor visible position %d out of bounds (cursor=%d, scroll=%d, vis=%d)",
				i+1, visPos, m.scroller.CursorLine(), m.scroller.Scroll(), m.visibleLines())
		}
	}

	if m.scroller.CursorLine() != 0 {
		t.Fatalf("expected cursorLine=0 at top, got %d", m.scroller.CursorLine())
	}
	if m.scroller.Scroll() != 0 {
		t.Fatalf("expected scroll=0 at top, got %d", m.scroller.Scroll())
	}
}
