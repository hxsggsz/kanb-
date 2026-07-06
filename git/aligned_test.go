package git

import (
	"testing"
)

func TestUnifiedAlignerAlignsContextLines(t *testing.T) {
	hunks := []Hunk{{
		OldStart: 1, OldCount: 2, NewStart: 1, NewCount: 2,
		Lines: []Line{
			{Type: LineContext, OldLineNum: 1, NewLineNum: 1, Content: "hello"},
			{Type: LineContext, OldLineNum: 2, NewLineNum: 2, Content: "world"},
		},
	}}
	a := &UnifiedAligner{}
	result := a.Align(hunks)
	if len(result) != 1 {
		t.Fatalf("expected 1 hunk, got %d", len(result))
	}
	if len(result[0].Lines) != 2 {
		t.Fatalf("expected 2 lines, got %d", len(result[0].Lines))
	}
	for _, l := range result[0].Lines {
		if l.OldContent != l.NewContent {
			t.Errorf("expected matching content, got left=%q right=%q", l.OldContent, l.NewContent)
		}
	}
}

func TestUnifiedAlignerDeletedLines(t *testing.T) {
	hunks := []Hunk{{
		OldStart: 1, OldCount: 2, NewStart: 1, NewCount: 0,
		Lines: []Line{
			{Type: LineDeleted, OldLineNum: 1, NewLineNum: 0, Content: "gone1"},
			{Type: LineDeleted, OldLineNum: 2, NewLineNum: 0, Content: "gone2"},
		},
	}}
	a := &UnifiedAligner{}
	result := a.Align(hunks)
	if len(result[0].Lines) != 2 {
		t.Fatalf("expected 2 lines, got %d", len(result[0].Lines))
	}
	for _, l := range result[0].Lines {
		if l.OldContent == "" || l.NewContent != "" {
			t.Errorf("expected old content only, got left=%q right=%q", l.OldContent, l.NewContent)
		}
	}
}

func TestUnifiedAlignerAddedLines(t *testing.T) {
	hunks := []Hunk{{
		OldStart: 1, OldCount: 0, NewStart: 1, NewCount: 2,
		Lines: []Line{
			{Type: LineAdded, OldLineNum: 0, NewLineNum: 1, Content: "new1"},
			{Type: LineAdded, OldLineNum: 0, NewLineNum: 2, Content: "new2"},
		},
	}}
	a := &UnifiedAligner{}
	result := a.Align(hunks)
	if len(result[0].Lines) != 2 {
		t.Fatalf("expected 2 lines, got %d", len(result[0].Lines))
	}
	for _, l := range result[0].Lines {
		if l.OldContent != "" || l.NewContent == "" {
			t.Errorf("expected new content only, got left=%q right=%q", l.OldContent, l.NewContent)
		}
	}
}

func TestUnifiedAlignerMixedChanges(t *testing.T) {
	hunks := []Hunk{{
		OldStart: 1, OldCount: 3, NewStart: 1, NewCount: 3,
		Lines: []Line{
			{Type: LineContext, OldLineNum: 1, NewLineNum: 1, Content: "stay"},
			{Type: LineDeleted, OldLineNum: 2, NewLineNum: 0, Content: "old"},
			{Type: LineAdded, OldLineNum: 0, NewLineNum: 2, Content: "new"},
			{Type: LineContext, OldLineNum: 3, NewLineNum: 3, Content: "keep"},
		},
	}}
	a := &UnifiedAligner{}
	result := a.Align(hunks)
	if len(result[0].Lines) != 4 {
		t.Fatalf("expected 4 lines, got %d", len(result[0].Lines))
	}
	if result[0].Lines[1].OldContent != "old" || result[0].Lines[1].NewContent != "" {
		t.Errorf("line 1: expected old='old' new='', got old=%q new=%q", result[0].Lines[1].OldContent, result[0].Lines[1].NewContent)
	}
	if result[0].Lines[2].OldContent != "" || result[0].Lines[2].NewContent != "new" {
		t.Errorf("line 2: expected old='' new='new', got old=%q new=%q", result[0].Lines[2].OldContent, result[0].Lines[2].NewContent)
	}
}
