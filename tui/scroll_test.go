package tui

import (
	"fmt"
	"testing"

	tea "charm.land/bubbletea/v2"
	"kanba/git"
)

// TestCursorStopsAtReturnZero reproduces the exact scenario the user reports:
// viewing a diff with cursor stuck at a specific line instead of reaching the end.
func TestCursorStopsAtReturnZero(t *testing.T) {
	// Reproduce the model.go diff structure:
	// Hunk 1: struct field changes + added totalLines() function — 40 body lines
	// Hunk 2: removed maxScroll() function — 19 body lines
	h1Lines := []git.Line{
		// context:   ) and blank
		{Type: git.LineContext, OldLineNum: 30, NewLineNum: 30, Content: ")"},
		{Type: git.LineContext, OldLineNum: 30, NewLineNum: 30, Content: ""},
		// context: type model struct {
		{Type: git.LineContext, OldLineNum: 31, NewLineNum: 31, Content: "type model struct {"},
		// deleted: old fields (8)
		{Type: git.LineDeleted, OldLineNum: 32, NewLineNum: 0, Content: "diffs   []git.FileDiff"},
		{Type: git.LineDeleted, OldLineNum: 33, NewLineNum: 0, Content: "fileIdx int"},
		{Type: git.LineDeleted, OldLineNum: 34, NewLineNum: 0, Content: "scroll  int"},
		{Type: git.LineDeleted, OldLineNum: 35, NewLineNum: 0, Content: "screen  screen"},
		{Type: git.LineDeleted, OldLineNum: 36, NewLineNum: 0, Content: "loading bool"},
		{Type: git.LineDeleted, OldLineNum: 37, NewLineNum: 0, Content: "err     error"},
		{Type: git.LineDeleted, OldLineNum: 38, NewLineNum: 0, Content: "width   int"},
		{Type: git.LineDeleted, OldLineNum: 39, NewLineNum: 0, Content: "height  int"},
		// added: new fields (9)
		{Type: git.LineAdded, OldLineNum: 0, NewLineNum: 32, Content: "diffs      []git.FileDiff"},
		{Type: git.LineAdded, OldLineNum: 0, NewLineNum: 33, Content: "fileIdx    int"},
		{Type: git.LineAdded, OldLineNum: 0, NewLineNum: 34, Content: "scroll     int"},
		{Type: git.LineAdded, OldLineNum: 0, NewLineNum: 35, Content: "cursorLine int"},
		{Type: git.LineAdded, OldLineNum: 0, NewLineNum: 36, Content: "screen     screen"},
		{Type: git.LineAdded, OldLineNum: 0, NewLineNum: 37, Content: "loading    bool"},
		{Type: git.LineAdded, OldLineNum: 0, NewLineNum: 38, Content: "err        error"},
		{Type: git.LineAdded, OldLineNum: 0, NewLineNum: 39, Content: "width      int"},
		{Type: git.LineAdded, OldLineNum: 0, NewLineNum: 40, Content: "height     int"},
		// context: blank, repoPath, gitArgs, }
		{Type: git.LineContext, OldLineNum: 40, NewLineNum: 41, Content: ""},
		{Type: git.LineContext, OldLineNum: 41, NewLineNum: 42, Content: "repoPath string"},
		{Type: git.LineContext, OldLineNum: 42, NewLineNum: 43, Content: "gitArgs  []string"},
		{Type: git.LineContext, OldLineNum: 43, NewLineNum: 44, Content: "}"},
		// context: blank
		{Type: git.LineContext, OldLineNum: 44, NewLineNum: 45, Content: ""},
		// added: totalLines function (11 lines including the return 0)
		{Type: git.LineAdded, OldLineNum: 0, NewLineNum: 46, Content: "func (m *model) totalLines() int {"},
		{Type: git.LineAdded, OldLineNum: 0, NewLineNum: 47, Content: "if len(m.diffs) == 0 {"},
		{Type: git.LineAdded, OldLineNum: 0, NewLineNum: 48, Content: "return 0"},
		{Type: git.LineAdded, OldLineNum: 0, NewLineNum: 49, Content: "}"},
		{Type: git.LineAdded, OldLineNum: 0, NewLineNum: 50, Content: "f := m.diffs[m.fileIdx]"},
		{Type: git.LineAdded, OldLineNum: 0, NewLineNum: 51, Content: "total := 0"},
		{Type: git.LineAdded, OldLineNum: 0, NewLineNum: 52, Content: "for _, h := range f.Hunks {"},
		{Type: git.LineAdded, OldLineNum: 0, NewLineNum: 53, Content: "total += hunkHeaderLines + len(h.Lines)"},
		{Type: git.LineAdded, OldLineNum: 0, NewLineNum: 54, Content: "}"},
		{Type: git.LineAdded, OldLineNum: 0, NewLineNum: 55, Content: "return total"},
		{Type: git.LineAdded, OldLineNum: 0, NewLineNum: 56, Content: "}"},
		// context: New, return, fields, }
		{Type: git.LineContext, OldLineNum: 45, NewLineNum: 57, Content: ""},
		{Type: git.LineContext, OldLineNum: 46, NewLineNum: 58, Content: "func New(repoPath string, gitArgs []string) tea.Model {"},
		{Type: git.LineContext, OldLineNum: 47, NewLineNum: 59, Content: "return &model{"},
	}

	// Hunk 2: context + deleted maxScroll (19 body lines)
	h2Lines := []git.Line{
		{Type: git.LineContext, OldLineNum: 57, NewLineNum: 70, Content: "func (m *model) visibleLines() int {"},
		{Type: git.LineContext, OldLineNum: 58, NewLineNum: 71, Content: "return m.height - (statusBarHeight + borderHeight)"},
		{Type: git.LineContext, OldLineNum: 59, NewLineNum: 72, Content: "}"},
		{Type: git.LineDeleted, OldLineNum: 60, NewLineNum: 0, Content: ""},
		{Type: git.LineDeleted, OldLineNum: 61, NewLineNum: 0, Content: "func (m *model) maxScroll() int {"},
		{Type: git.LineDeleted, OldLineNum: 62, NewLineNum: 0, Content: "if len(m.diffs) == 0 {"},
		{Type: git.LineDeleted, OldLineNum: 63, NewLineNum: 0, Content: "return 0"},
		{Type: git.LineDeleted, OldLineNum: 64, NewLineNum: 0, Content: "}"},
		{Type: git.LineDeleted, OldLineNum: 65, NewLineNum: 0, Content: "f := m.diffs[m.fileIdx]"},
		{Type: git.LineDeleted, OldLineNum: 66, NewLineNum: 0, Content: "totalLines := 0"},
		{Type: git.LineDeleted, OldLineNum: 67, NewLineNum: 0, Content: "for _, h := range f.Hunks {"},
		{Type: git.LineDeleted, OldLineNum: 68, NewLineNum: 0, Content: "totalLines += hunkHeaderLines + len(h.Lines)"},
		{Type: git.LineDeleted, OldLineNum: 69, NewLineNum: 0, Content: "}"},
		{Type: git.LineDeleted, OldLineNum: 70, NewLineNum: 0, Content: "max := totalLines - m.visibleLines()"},
		{Type: git.LineDeleted, OldLineNum: 71, NewLineNum: 0, Content: "if max < 0 {"},
		{Type: git.LineDeleted, OldLineNum: 72, NewLineNum: 0, Content: "return 0"},
		{Type: git.LineDeleted, OldLineNum: 73, NewLineNum: 0, Content: "}"},
		{Type: git.LineDeleted, OldLineNum: 74, NewLineNum: 0, Content: "return max"},
		{Type: git.LineDeleted, OldLineNum: 75, NewLineNum: 0, Content: "}"},
	}

	f := git.FileDiff{
		NewPath: "tui/model.go",
		Status:  "M",
		Hunks: []git.Hunk{
			{Header: "@@ -29,19 +29,32 @@ const (", Lines: h1Lines},
			{Header: "@@ -57,19 +70,3 @@ func (m *model) Init() tea.Cmd {", Lines: h2Lines},
		},
	}

	totalLines := 1 + len(h1Lines) + 1 + len(h2Lines)
	t.Logf("h1Lines=%d, h2Lines=%d, totalLines=%d", len(h1Lines), len(h2Lines), totalLines)

	for height := 12; height <= 40; height++ {
		t.Run(fmt.Sprintf("height=%d", height), func(t *testing.T) {
			m := &model{
				diffs:      []git.FileDiff{f},
				fileIdx:    0,
				cursorLine: 0,
				scroll:     0,
				width:      80,
				height:     height,
			}

			vis := m.visibleLines()
			if vis <= 0 {
				t.Skip("vis <= 0")
			}

			// Navigate all the way down
			for i := 0; i < totalLines-1; i++ {
				m.handleDiffKeys(tea.KeyPressMsg{Code: tea.KeyDown})
				m.renderFile(f, 80, m.visibleLines())

				visPos := m.cursorLine - m.scroll
				if visPos < 0 || visPos >= vis {
					t.Fatalf("DOWN step %d: cursorLine=%d outside view [scroll=%d, scroll+vis=%d) visPos=%d",
						i+1, m.cursorLine, m.scroll, m.scroll+vis, visPos)
				}
			}

			// Verify final state
			if m.cursorLine != totalLines-1 {
				t.Fatalf("expected cursorLine=%d (last line), got %d", totalLines-1, m.cursorLine)
			}

			// Verify totalLines() matches totalLines used for rendering
			if m.totalLines() != totalLines {
				t.Fatalf("m.totalLines()=%d != expected totalLines=%d", m.totalLines(), totalLines)
			}
		})
	}
}

func TestScrollForDifferentHeights(t *testing.T) {
	for height := 5; height <= 40; height++ {
		t.Run(fmt.Sprintf("height=%d", height), func(t *testing.T) {
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
				height:     height,
			}

			total := m.totalLines()
			vis := m.visibleLines()

			if vis <= 0 {
				t.Skip("visible lines <= 0")
			}

			// Navigate all the way down
			for i := 0; i < total-1; i++ {
				m.handleDiffKeys(tea.KeyPressMsg{Code: tea.KeyDown})
				m.renderFile(f, 80, m.visibleLines())

				visPos := m.cursorLine - m.scroll
				if visPos < 0 || visPos >= vis {
					t.Fatalf("DOWN: cursorLine=%d outside view [scroll=%d, scroll+vis=%d) visPos=%d",
						m.cursorLine, m.scroll, m.scroll+vis, visPos)
				}

				if m.scroll < 0 {
					t.Fatalf("DOWN: negative scroll=%d", m.scroll)
				}
			}

			if m.cursorLine != total-1 {
				t.Fatalf("expected cursor at last line %d, got %d", total-1, m.cursorLine)
			}

			// Navigate all the way up
			for i := 0; i < total-1; i++ {
				m.handleDiffKeys(tea.KeyPressMsg{Code: tea.KeyUp})
				m.renderFile(f, 80, m.visibleLines())

				visPos := m.cursorLine - m.scroll
				if visPos < 0 || visPos >= vis {
					t.Fatalf("UP: cursorLine=%d outside view [scroll=%d, scroll+vis=%d) visPos=%d",
						m.cursorLine, m.scroll, m.scroll+vis, visPos)
				}

				if m.scroll < 0 {
					t.Fatalf("UP: negative scroll=%d", m.scroll)
				}
			}

			if m.cursorLine != 0 {
				t.Fatalf("expected cursor at line 0, got %d", m.cursorLine)
			}
		})
	}
}

func TestScrollStallDetector(t *testing.T) {
	type testCase struct {
		name     string
		numLines int
		height   int
	}

	cases := []testCase{
		{"h24_n50", 50, 24},
		{"h24_n100", 100, 24},
		{"h24_n500", 500, 24},
		{"h20_n50", 50, 20},
		{"h15_n50", 50, 15},
		{"h12_n50", 50, 12},
		{"h10_n50", 50, 10},
		{"h8_n50", 50, 8},
		{"h24_n10", 10, 24},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			lines := make([]git.Line, tc.numLines)
			for i := range lines {
				lines[i] = git.Line{Type: git.LineContext, OldLineNum: i + 1, NewLineNum: i + 1, Content: "line"}
			}
			h := git.Hunk{
				Header: "@@ -1,N +1,N @@",
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
				height:     tc.height,
			}

			total := m.totalLines()
			vis := m.visibleLines()

			if vis <= 0 {
				t.Skip("vis <= 0")
			}

			maxScroll := total - vis
			if maxScroll < 0 {
				maxScroll = 0
			}

			scrollMargin := 8
			prevScroll := -1
			stallCount := 0

			for i := 0; i < total; i++ {
				if i > 0 {
					m.handleDiffKeys(tea.KeyPressMsg{Code: tea.KeyDown})
				}
				m.renderFile(f, 80, m.visibleLines())

				if m.scroll == prevScroll && m.cursorLine > m.scroll+vis-scrollMargin {
					if m.scroll < maxScroll {
						stallCount++
						if stallCount > 3 {
							t.Fatalf("SCROLL STALL at step cursorLine=%d scroll=%d prevScroll=%d maxScroll=%d vis=%d",
								m.cursorLine, m.scroll, prevScroll, maxScroll, vis)
						}
					}
				} else {
					if m.scroll != prevScroll {
						stallCount = 0
					}
				}

				prevScroll = m.scroll
			}
		})
	}
}
