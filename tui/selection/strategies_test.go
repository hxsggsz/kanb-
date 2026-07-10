package selection

import (
	"testing"
)

func TestCharacterStrategy_Select(t *testing.T) {
	cs := CharacterStrategy{}

	tests := []struct {
		name     string
		content  string
		startCol int
		endCol   int
		expected Range
	}{
		{
			name:     "pass-through start and end",
			content:  "hello world",
			startCol: 2,
			endCol:   7,
			expected: Range{StartLine: 0, StartCol: 2, EndLine: 0, EndCol: 7},
		},
		{
			name:     "empty range",
			content:  "hello",
			startCol: 3,
			endCol:   3,
			expected: Range{StartLine: 0, StartCol: 3, EndLine: 0, EndCol: 3},
		},
		{
			name:     "reversed range preserved",
			content:  "hello",
			startCol: 4,
			endCol:   1,
			expected: Range{StartLine: 0, StartCol: 4, EndLine: 0, EndCol: 1},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := cs.Select(tt.content, tt.startCol, tt.endCol)
			if result != tt.expected {
				t.Errorf("got %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestIsWordChar(t *testing.T) {
	tests := []struct {
		r        rune
		expected bool
	}{
		{'a', true},
		{'Z', true},
		{'0', true},
		{'9', true},
		{'_', true},
		{' ', false},
		{'.', false},
		{'-', false},
		{'!', false},
		{'@', false},
		{'ç', true},
		{'é', true},
		{'中', true},
	}

	for _, tt := range tests {
		t.Run(string(tt.r), func(t *testing.T) {
			if got := IsWordChar(tt.r); got != tt.expected {
				t.Errorf("IsWordChar(%q) = %v, want %v", tt.r, got, tt.expected)
			}
		})
	}
}

func TestFindWordBoundaries(t *testing.T) {
	tests := []struct {
		name          string
		line          string
		col           int
		expectedStart int
		expectedEnd   int
	}{
		{
			name:          "middle of word",
			line:          "hello world",
			col:           2,
			expectedStart: 0,
			expectedEnd:   4,
		},
		{
			name:          "start of word",
			line:          "hello world",
			col:           0,
			expectedStart: 0,
			expectedEnd:   4,
		},
		{
			name:          "end of word",
			line:          "hello world",
			col:           4,
			expectedStart: 0,
			expectedEnd:   4,
		},
		{
			name:          "on space",
			line:          "hello world",
			col:           5,
			expectedStart: 5,
			expectedEnd:   5,
		},
		{
			name:          "single char word",
			line:          "a b",
			col:           0,
			expectedStart: 0,
			expectedEnd:   0,
		},
		{
			name:          "end of line",
			line:          "hello",
			col:           4,
			expectedStart: 0,
			expectedEnd:   4,
		},
		{
			name:          "empty line",
			line:          "",
			col:           0,
			expectedStart: 0,
			expectedEnd:   0,
		},
		{
			name:          "underscores in word",
			line:          "foo_bar baz",
			col:           2,
			expectedStart: 0,
			expectedEnd:   6,
		},
		{
			name:          "underscore at cursor",
			line:          "foo_bar baz",
			col:           3,
			expectedStart: 0,
			expectedEnd:   6,
		},
		{
			name:          "UTF-8 characters",
			line:          "café résumé",
			col:           2,
			expectedStart: 0,
			expectedEnd:   3,
		},
		{
			name:          "UTF-8 middle of word",
			line:          "café résumé",
			col:           8,
			expectedStart: 5,
			expectedEnd:   10,
		},
		{
			name:          "col beyond line length",
			line:          "hello",
			col:           10,
			expectedStart: 0,
			expectedEnd:   4,
		},
		{
			name:          "negative col",
			line:          "hello",
			col:           -1,
			expectedStart: 0,
			expectedEnd:   4,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			start, end := findWordBoundaries(tt.line, tt.col)
			if start != tt.expectedStart || end != tt.expectedEnd {
				t.Errorf("findWordBoundaries(%q, %d) = (%d, %d), want (%d, %d)",
					tt.line, tt.col, start, end, tt.expectedStart, tt.expectedEnd)
			}
		})
	}
}

func TestWordStrategy_Select(t *testing.T) {
	ws := WordStrategy{}

	tests := []struct {
		name     string
		content  string
		startCol int
		endCol   int
		expected Range
	}{
		{
			name:     "selects whole word on double-click",
			content:  "hello world",
			startCol: 2,
			endCol:   2,
			expected: Range{StartLine: 0, StartCol: 0, EndLine: 0, EndCol: 4},
		},
		{
			name:     "selects multiple words",
			content:  "hello world",
			startCol: 0,
			endCol:   7,
			expected: Range{StartLine: 0, StartCol: 0, EndLine: 0, EndCol: 10},
		},
		{
			name:     "on space returns single position",
			content:  "hello world",
			startCol: 5,
			endCol:   5,
			expected: Range{StartLine: 0, StartCol: 5, EndLine: 0, EndCol: 5},
		},
		{
			name:     "multi-byte UTF-8 content expands to full word",
			content:  "éééé",
			startCol: 3,
			endCol:   3,
			expected: Range{StartLine: 0, StartCol: 0, EndLine: 0, EndCol: 3},
		},
		{
			name:     "punctuation between words breaks selection",
			content:  "foo.bar",
			startCol: 3,
			endCol:   3,
			expected: Range{StartLine: 0, StartCol: 3, EndLine: 0, EndCol: 3},
		},
		{
			name:     "mixed alphanumeric is single word",
			content:  "foo123bar",
			startCol: 4,
			endCol:   4,
			expected: Range{StartLine: 0, StartCol: 0, EndLine: 0, EndCol: 8},
		},
		{
			name:     "reversed drag on multi-byte content",
			content:  "éééé",
			startCol: 3,
			endCol:   1,
			expected: Range{StartLine: 0, StartCol: 0, EndLine: 0, EndCol: 3},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ws.Select(tt.content, tt.startCol, tt.endCol)
			if result != tt.expected {
				t.Errorf("got %v, want %v", result, tt.expected)
			}
		})
	}
}
