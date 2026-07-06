package tui

import (
	"fmt"
	"testing"

	tea "charm.land/bubbletea/v2"
	"kanba/git"
)

func TestCursorStopsAtReturnZero(t *testing.T) {
	h1Lines := []git.AlignedLine{
		{Kind: git.KindContext, OldLineNum: 30, NewLineNum: 30, OldContent: ")", NewContent: ")"},
		{Kind: git.KindContext, OldLineNum: 30, NewLineNum: 30, OldContent: "", NewContent: ""},
		{Kind: git.KindContext, OldLineNum: 31, NewLineNum: 31, OldContent: "type model struct {", NewContent: "type model struct {"},
		{Kind: git.KindDeleted, OldLineNum: 32, NewLineNum: 0, OldContent: "diffs   []git.FileDiff"},
		{Kind: git.KindDeleted, OldLineNum: 33, NewLineNum: 0, OldContent: "fileIdx int"},
		{Kind: git.KindDeleted, OldLineNum: 34, NewLineNum: 0, OldContent: "scroll  int"},
		{Kind: git.KindDeleted, OldLineNum: 35, NewLineNum: 0, OldContent: "screen  screen"},
		{Kind: git.KindDeleted, OldLineNum: 36, NewLineNum: 0, OldContent: "loading bool"},
		{Kind: git.KindDeleted, OldLineNum: 37, NewLineNum: 0, OldContent: "err     error"},
		{Kind: git.KindDeleted, OldLineNum: 38, NewLineNum: 0, OldContent: "width   int"},
		{Kind: git.KindDeleted, OldLineNum: 39, NewLineNum: 0, OldContent: "height  int"},
		{Kind: git.KindAdded, OldLineNum: 0, NewLineNum: 32, NewContent: "diffs      []git.FileDiff"},
		{Kind: git.KindAdded, OldLineNum: 0, NewLineNum: 33, NewContent: "fileIdx    int"},
		{Kind: git.KindAdded, OldLineNum: 0, NewLineNum: 34, NewContent: "scroll     int"},
		{Kind: git.KindAdded, OldLineNum: 0, NewLineNum: 35, NewContent: "cursorLine int"},
		{Kind: git.KindAdded, OldLineNum: 0, NewLineNum: 36, NewContent: "screen     screen"},
		{Kind: git.KindAdded, OldLineNum: 0, NewLineNum: 37, NewContent: "loading    bool"},
		{Kind: git.KindAdded, OldLineNum: 0, NewLineNum: 38, NewContent: "err        error"},
		{Kind: git.KindAdded, OldLineNum: 0, NewLineNum: 39, NewContent: "width      int"},
		{Kind: git.KindAdded, OldLineNum: 0, NewLineNum: 40, NewContent: "height     int"},
		{Kind: git.KindContext, OldLineNum: 40, NewLineNum: 41, OldContent: "", NewContent: ""},
		{Kind: git.KindContext, OldLineNum: 41, NewLineNum: 42, OldContent: "repoPath string", NewContent: "repoPath string"},
		{Kind: git.KindContext, OldLineNum: 42, NewLineNum: 43, OldContent: "gitArgs  []string", NewContent: "gitArgs  []string"},
		{Kind: git.KindContext, OldLineNum: 43, NewLineNum: 44, OldContent: "}", NewContent: "}"},
		{Kind: git.KindContext, OldLineNum: 44, NewLineNum: 45, OldContent: "", NewContent: ""},
		{Kind: git.KindAdded, OldLineNum: 0, NewLineNum: 46, NewContent: "func (m *model) totalLines() int {"},
		{Kind: git.KindAdded, OldLineNum: 0, NewLineNum: 47, NewContent: "if len(m.diffs) == 0 {"},
		{Kind: git.KindAdded, OldLineNum: 0, NewLineNum: 48, NewContent: "return 0"},
		{Kind: git.KindAdded, OldLineNum: 0, NewLineNum: 49, NewContent: "}"},
		{Kind: git.KindAdded, OldLineNum: 0, NewLineNum: 50, NewContent: "f := m.diffs[m.fileIdx]"},
		{Kind: git.KindAdded, OldLineNum: 0, NewLineNum: 51, NewContent: "total := 0"},
		{Kind: git.KindAdded, OldLineNum: 0, NewLineNum: 52, NewContent: "for _, h := range f.Hunks {"},
		{Kind: git.KindAdded, OldLineNum: 0, NewLineNum: 53, NewContent: "total += hunkHeaderLines + len(h.Lines)"},
		{Kind: git.KindAdded, OldLineNum: 0, NewLineNum: 54, NewContent: "}"},
		{Kind: git.KindAdded, OldLineNum: 0, NewLineNum: 55, NewContent: "return total"},
		{Kind: git.KindAdded, OldLineNum: 0, NewLineNum: 56, NewContent: "}"},
		{Kind: git.KindContext, OldLineNum: 45, NewLineNum: 57, OldContent: "", NewContent: ""},
		{Kind: git.KindContext, OldLineNum: 46, NewLineNum: 58, OldContent: "func New(repoPath string, gitArgs []string) tea.Model {", NewContent: "func New(repoPath string, gitArgs []string) tea.Model {"},
		{Kind: git.KindContext, OldLineNum: 47, NewLineNum: 59, OldContent: "return &model{", NewContent: "return &model{"},
	}

	h2Lines := []git.AlignedLine{
		{Kind: git.KindContext, OldLineNum: 57, NewLineNum: 70, OldContent: "func (m *model) visibleLines() int {", NewContent: "func (m *model) visibleLines() int {"},
		{Kind: git.KindContext, OldLineNum: 58, NewLineNum: 71, OldContent: "return m.height - (statusBarHeight + borderHeight)", NewContent: "return m.height - (statusBarHeight + borderHeight)"},
		{Kind: git.KindContext, OldLineNum: 59, NewLineNum: 72, OldContent: "}", NewContent: "}"},
		{Kind: git.KindDeleted, OldLineNum: 60, NewLineNum: 0, OldContent: ""},
		{Kind: git.KindDeleted, OldLineNum: 61, NewLineNum: 0, OldContent: "func (m *model) maxScroll() int {"},
		{Kind: git.KindDeleted, OldLineNum: 62, NewLineNum: 0, OldContent: "if len(m.diffs) == 0 {"},
		{Kind: git.KindDeleted, OldLineNum: 63, NewLineNum: 0, OldContent: "return 0"},
		{Kind: git.KindDeleted, OldLineNum: 64, NewLineNum: 0, OldContent: "}"},
		{Kind: git.KindDeleted, OldLineNum: 65, NewLineNum: 0, OldContent: "f := m.diffs[m.fileIdx]"},
		{Kind: git.KindDeleted, OldLineNum: 66, NewLineNum: 0, OldContent: "totalLines := 0"},
		{Kind: git.KindDeleted, OldLineNum: 67, NewLineNum: 0, OldContent: "for _, h := range f.Hunks {"},
		{Kind: git.KindDeleted, OldLineNum: 68, NewLineNum: 0, OldContent: "totalLines += hunkHeaderLines + len(h.Lines)"},
		{Kind: git.KindDeleted, OldLineNum: 69, NewLineNum: 0, OldContent: "}"},
		{Kind: git.KindDeleted, OldLineNum: 70, NewLineNum: 0, OldContent: "max := totalLines - m.visibleLines()"},
		{Kind: git.KindDeleted, OldLineNum: 71, NewLineNum: 0, OldContent: "if max < 0 {"},
		{Kind: git.KindDeleted, OldLineNum: 72, NewLineNum: 0, OldContent: "return 0"},
		{Kind: git.KindDeleted, OldLineNum: 73, NewLineNum: 0, OldContent: "}"},
		{Kind: git.KindDeleted, OldLineNum: 74, NewLineNum: 0, OldContent: "return max"},
		{Kind: git.KindDeleted, OldLineNum: 75, NewLineNum: 0, OldContent: "}"},
	}

	f := git.SideBySideDiff{
		NewPath: "tui/model.go",
		Status:  "M",
		Hunks: []git.AlignedHunk{
			{Header: "@@ -29,19 +29,32 @@ const (", Lines: h1Lines},
			{Header: "@@ -57,19 +70,3 @@ func (m *model) Init() tea.Cmd {", Lines: h2Lines},
		},
	}

	totalLines := 1 + len(h1Lines) + 1 + len(h2Lines)
	t.Logf("h1Lines=%d, h2Lines=%d, totalLines=%d", len(h1Lines), len(h2Lines), totalLines)

	for height := 12; height <= 40; height++ {
		t.Run(fmt.Sprintf("height=%d", height), func(t *testing.T) {
			m := &model{
				diffs:      []git.SideBySideDiff{f},
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

			for i := 0; i < totalLines-1; i++ {
				m.handleDiffKeys(tea.KeyPressMsg{Code: tea.KeyDown})
				m.renderFile(f, 80, m.visibleLines())

				visPos := m.cursorLine - m.scroll
				if visPos < 0 || visPos >= vis {
					t.Fatalf("DOWN step %d: cursorLine=%d outside view [scroll=%d, scroll+vis=%d) visPos=%d",
						i+1, m.cursorLine, m.scroll, m.scroll+vis, visPos)
				}
			}

			if m.cursorLine != totalLines-1 {
				t.Fatalf("expected cursorLine=%d (last line), got %d", totalLines-1, m.cursorLine)
			}

			if m.totalLines() != totalLines {
				t.Fatalf("m.totalLines()=%d != expected totalLines=%d", m.totalLines(), totalLines)
			}
		})
	}
}

func TestScrollForDifferentHeights(t *testing.T) {
	for height := 5; height <= 40; height++ {
		t.Run(fmt.Sprintf("height=%d", height), func(t *testing.T) {
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
				diffs:      []git.SideBySideDiff{f},
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
			lines := make([]git.AlignedLine, tc.numLines)
			for i := range lines {
				lines[i] = git.AlignedLine{Kind: git.KindContext, OldLineNum: i + 1, NewLineNum: i + 1, OldContent: "line", NewContent: "line"}
			}
			h := git.AlignedHunk{
				Header: "@@ -1,N +1,N @@",
				Lines:  lines,
			}
			f := git.SideBySideDiff{
				NewPath: "test.txt",
				Status:  "M",
				Hunks:   []git.AlignedHunk{h},
			}

			m := &model{
				diffs:      []git.SideBySideDiff{f},
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
