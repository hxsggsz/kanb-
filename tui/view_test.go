package tui

import (
	"testing"

	tea "charm.land/bubbletea/v2"
	"kanba/git"
)

func TestCursorAtEndOfFile(t *testing.T) {
	h := git.Hunk{
		Header: "@@ -1,5 +1,6 @@",
		Lines: []git.Line{
			{Type: git.LineContext, OldLineNum: 1, NewLineNum: 1, Content: "line1"},
			{Type: git.LineContext, OldLineNum: 2, NewLineNum: 2, Content: "line2"},
			{Type: git.LineContext, OldLineNum: 3, NewLineNum: 3, Content: "line3"},
			{Type: git.LineContext, OldLineNum: 4, NewLineNum: 4, Content: "line4"},
			{Type: git.LineAdded, OldLineNum: 0, NewLineNum: 5, Content: "line5new"},
		},
	}
	f := git.FileDiff{
		NewPath: "test.txt",
		Status:  "M",
		Hunks:   []git.Hunk{h},
	}

	m := &model{
		diffs:      []git.FileDiff{f},
		fileIdx:    0,
		cursorLine: 0,
		scroll:     0,
		width:      80,
		height:     24,
	}

	total := m.totalLines()
	t.Logf("totalLines: %d", total)

	nav := func(steps int) {
		for i := 0; i < steps; i++ {
			m.handleDiffKeys(tea.KeyPressMsg{Code: tea.KeyDown})
		}
	}

	nav(5)
	if m.cursorLine != 5 {
		t.Fatalf("expected cursorLine=5, got %d", m.cursorLine)
	}

	nav(1)
	if m.cursorLine != 5 {
		t.Fatalf("cursor moved past end: expected 5, got %d", m.cursorLine)
	}

	nav(100)
	if m.cursorLine != 5 {
		t.Fatalf("cursor moved past end after 100 presses: expected 5, got %d", m.cursorLine)
	}

	m.renderFile(f, 80, m.visibleLines())
	if m.cursorLine != 5 {
		t.Fatalf("renderFile changed cursorLine: expected 5, got %d", m.cursorLine)
	}
	if m.cursorLine >= m.totalLines() {
		t.Fatalf("cursorLine %d >= totalLines %d after renderFile", m.cursorLine, m.totalLines())
	}
}

func TestScrollAdvancesWithCursor(t *testing.T) {
	lines := make([]git.Line, 200)
	for i := range lines {
		lines[i] = git.Line{Type: git.LineContext, OldLineNum: i + 1, NewLineNum: i + 1, Content: "line"}
	}
	h := git.Hunk{
		Header: "@@ -1,200 +1,200 @@",
		Lines:  lines,
	}
	f := git.FileDiff{
		NewPath: "test.txt",
		Status:  "M",
		Hunks:   []git.Hunk{h},
	}

	m := &model{
		diffs:      []git.FileDiff{f},
		fileIdx:    0,
		cursorLine: 0,
		scroll:     0,
		width:      80,
		height:     24,
	}

	total := m.totalLines()
	t.Logf("totalLines: %d, expected: %d", total, 201)

	vis := m.visibleLines()

	for i := 0; i < total-1; i++ {
		m.handleDiffKeys(tea.KeyPressMsg{Code: tea.KeyDown})

		m.renderFile(f, 80, m.visibleLines())

		visPos := m.cursorLine - m.scroll
		if visPos < 0 || visPos >= vis {
			t.Fatalf("step %d: cursor visible position %d (cursorLine=%d, scroll=%d) out of view [0, %d)",
				i+1, visPos, m.cursorLine, m.scroll, vis)
		}

		if m.cursorLine >= vis && m.scroll == 0 {
			t.Fatalf("step %d: scroll is 0 but cursorLine=%d is past visible area (vis=%d)",
				i+1, m.cursorLine, vis)
		}
	}

	t.Logf("final state: cursorLine=%d, scroll=%d, total=%d, vis=%d",
		m.cursorLine, m.scroll, total, vis)
	if m.cursorLine != total-1 {
		t.Fatalf("expected cursorLine=%d, got %d", total-1, m.cursorLine)
	}
}

func TestScrollDoesNotGetStuck(t *testing.T) {
	lines := make([]git.Line, 100)
	for i := range lines {
		lines[i] = git.Line{Type: git.LineContext, OldLineNum: i + 1, NewLineNum: i + 1, Content: "line"}
	}
	h := git.Hunk{
		Header: "@@ -1,100 +1,100 @@",
		Lines:  lines,
	}
	f := git.FileDiff{
		NewPath: "test.txt",
		Status:  "M",
		Hunks:   []git.Hunk{h},
	}

	m := &model{
		diffs:      []git.FileDiff{f},
		fileIdx:    0,
		cursorLine: 0,
		scroll:     0,
		width:      80,
		height:     24,
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

		if m.cursorLine >= vis {
			if m.scroll == lastScroll {
				scrollStallCount++
			} else {
				scrollStallCount = 0
			}

			if scrollStallCount > 5 && m.scroll < maxScroll {
				t.Fatalf("step %d: scroll stuck at %d for %d steps, maxScroll=%d, cursorLine=%d",
					i, m.scroll, scrollStallCount, maxScroll, m.cursorLine)
			}
		}

		lastScroll = m.scroll
	}

	t.Logf("final: cursorLine=%d, scroll=%d, maxScroll=%d",
		m.cursorLine, m.scroll, maxScroll)
}

func TestScrollKeepsCursorInViewForLargeFile(t *testing.T) {
	lines := make([]git.Line, 500)
	for i := range lines {
		lines[i] = git.Line{Type: git.LineContext, OldLineNum: i + 1, NewLineNum: i + 1, Content: "line"}
	}
	h := git.Hunk{
		Header: "@@ -1,500 +1,500 @@",
		Lines:  lines,
	}
	f := git.FileDiff{
		NewPath: "test.txt",
		Status:  "M",
		Hunks:   []git.Hunk{h},
	}

	m := &model{
		diffs:      []git.FileDiff{f},
		fileIdx:    0,
		cursorLine: 0,
		scroll:     0,
		width:      80,
		height:     24,
	}

	total := m.totalLines()
	vis := m.visibleLines()

	for i := 0; i < total-1; i++ {
		m.handleDiffKeys(tea.KeyPressMsg{Code: tea.KeyDown})
		m.renderFile(f, 80, m.visibleLines())

		if m.cursorLine < m.scroll || m.cursorLine >= m.scroll+vis {
			t.Fatalf("cursorLine=%d outside view [scroll=%d, scroll+vis=%d) at step %d",
				m.cursorLine, m.scroll, m.scroll+vis, i+1)
		}
	}

	if m.cursorLine != total-1 {
		t.Fatalf("expected cursor at last line %d, got %d", total-1, m.cursorLine)
	}
}

func TestScrollUpFromBottom(t *testing.T) {
	lines := make([]git.Line, 100)
	for i := range lines {
		lines[i] = git.Line{Type: git.LineContext, OldLineNum: i + 1, NewLineNum: i + 1, Content: "line"}
	}
	h := git.Hunk{
		Header: "@@ -1,100 +1,100 @@",
		Lines:  lines,
	}
	f := git.FileDiff{
		NewPath: "test.txt",
		Status:  "M",
		Hunks:   []git.Hunk{h},
	}

	m := &model{
		diffs:      []git.FileDiff{f},
		fileIdx:    0,
		cursorLine: 0,
		scroll:     0,
		width:      80,
		height:     24,
	}

	total := m.totalLines()

	// Go to bottom
	for i := 0; i < total-1; i++ {
		m.handleDiffKeys(tea.KeyPressMsg{Code: tea.KeyDown})
	}
	m.renderFile(f, 80, m.visibleLines())

	t.Logf("at bottom: cursorLine=%d, scroll=%d", m.cursorLine, m.scroll)

	// Go back up and verify scroll decreases
	for i := 0; i < total-1; i++ {
		m.handleDiffKeys(tea.KeyPressMsg{Code: tea.KeyUp})
		m.renderFile(f, 80, m.visibleLines())

		if m.scroll < 0 {
			t.Fatalf("negative scroll at step %d: %d", i+1, m.scroll)
		}
	}

	if m.cursorLine != 0 {
		t.Fatalf("expected cursor at line 0, got %d", m.cursorLine)
	}
	if m.scroll != 0 {
		t.Fatalf("expected scroll=0 at top, got %d", m.scroll)
	}
}

func TestScrollHysteresisCursorVisiblePosition(t *testing.T) {
	lines := make([]git.Line, 50)
	for i := range lines {
		lines[i] = git.Line{Type: git.LineContext, OldLineNum: i + 1, NewLineNum: i + 1, Content: "line"}
	}
	h := git.Hunk{
		Header: "@@ -1,50 +1,50 @@",
		Lines:  lines,
	}
	f := git.FileDiff{
		NewPath: "test.txt",
		Status:  "M",
		Hunks:   []git.Hunk{h},
	}

	m := &model{
		diffs:      []git.FileDiff{f},
		fileIdx:    0,
		cursorLine: 0,
		scroll:     0,
		width:      80,
		height:     24,
	}

	total := m.totalLines()

	// Navigate all the way down, check cursor stays in view
	for i := 0; i < total-1; i++ {
		m.handleDiffKeys(tea.KeyPressMsg{Code: tea.KeyDown})
		m.renderFile(f, 80, m.visibleLines())
		visPos := m.cursorLine - m.scroll
		if visPos < 0 || visPos >= m.visibleLines() {
			t.Fatalf("DOWN step %d: cursor visible position %d out of bounds (cursor=%d, scroll=%d, vis=%d)",
				i+1, visPos, m.cursorLine, m.scroll, m.visibleLines())
		}
	}

	// Navigate all the way up, check cursor stays in view
	for i := 0; i < total-1; i++ {
		m.handleDiffKeys(tea.KeyPressMsg{Code: tea.KeyUp})
		m.renderFile(f, 80, m.visibleLines())
		visPos := m.cursorLine - m.scroll
		if visPos < 0 || visPos >= m.visibleLines() {
			t.Fatalf("UP step %d: cursor visible position %d out of bounds (cursor=%d, scroll=%d, vis=%d)",
				i+1, visPos, m.cursorLine, m.scroll, m.visibleLines())
		}
	}

	if m.cursorLine != 0 {
		t.Fatalf("expected cursorLine=0 at top, got %d", m.cursorLine)
	}
	if m.scroll != 0 {
		t.Fatalf("expected scroll=0 at top, got %d", m.scroll)
	}
}
