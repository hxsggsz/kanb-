package tui

import (
	"strings"
	"testing"

	"kanba/git"
)

func TestRenderAlignedLinePreservesANSI(t *testing.T) {
	sh := NewSyntaxHighlighter()
	ln := git.AlignedLine{
		OldLineNum: 1,
		NewLineNum: 1,
		OldContent: "func main() {",
		NewContent: "func main() {",
		Kind:       git.KindContext,
	}
	fmtr := defaultFormatters[ln.Kind]
	result := renderAlignedLine(fmtr, ln, 80, false, sh, "main.go", 0, false, RosePine)
	if !strings.Contains(result, "\x1b[") {
		t.Fatal("expected ANSI escape codes in rendered output")
	}
}

func TestRenderAlignedLineAddsBackground(t *testing.T) {
	sh := NewSyntaxHighlighter()
	ln := git.AlignedLine{
		NewLineNum: 1,
		NewContent: "func main() {",
		Kind:       git.KindAdded,
	}
	fmtr := defaultFormatters[ln.Kind]
	result := renderAlignedLine(fmtr, ln, 80, false, sh, "main.go", 0, false, RosePine)

	// RosePine.AddedBg = "#333c48" → 48;2;51;60;72
	if !strings.Contains(result, "48;2;51;60;72") {
		t.Fatalf("expected RosePine added background, got: %q", result)
	}

	sep := " │ "
	idx := strings.Index(result, sep)
	if idx < 0 {
		t.Fatalf("expected separator %q in result: %q", sep, result)
	}
	rightSide := result[idx+len(sep):]
	rightPrefix := "\x1b[38;2;156;207;216;48;2;51;60;72m"
	if !strings.HasPrefix(rightSide, rightPrefix) {
		t.Fatalf("expected right side to start with %s, got: %q", rightPrefix, rightSide[:50])
	}

	if !strings.Contains(result, "\x1b[1m") && !strings.Contains(result, "\x1b[3") {
		t.Fatalf("expected syntax highlighting (\\x1b[1m or \\x1b[3), got: %q", result)
	}
}

func TestRenderAlignedLineDeletedBackground(t *testing.T) {
	sh := NewSyntaxHighlighter()
	ln := git.AlignedLine{
		OldLineNum: 1,
		OldContent: "func main() {",
		Kind:       git.KindDeleted,
	}
	fmtr := defaultFormatters[ln.Kind]
	result := renderAlignedLine(fmtr, ln, 80, false, sh, "main.go", 0, false, RosePine)

	// RosePine.RemovedBg = "#312e3f" → 48;2;49;46;63
	if !strings.Contains(result, "48;2;49;46;63") {
		t.Fatalf("expected RosePine removed background, got: %q", result)
	}

	leftPrefix := "\x1b[38;2;144;140;170;48;2;49;46;63m"
	if !strings.HasPrefix(result, leftPrefix) {
		t.Fatalf("expected result to start with %s, got: %q", leftPrefix, result[:50])
	}
}

func TestRenderAlignedLineSinglePanel(t *testing.T) {
	sh := NewSyntaxHighlighter()
	ln := git.AlignedLine{
		NewLineNum: 1,
		NewContent: "func main() {",
		Kind:       git.KindAdded,
	}
	fmtr := defaultFormatters[ln.Kind]
	result := renderAlignedLine(fmtr, ln, 80, false, sh, "main.go", 0, true, RosePine)

	if strings.Contains(result, " │ ") {
		t.Fatal("single-panel result should not contain separator")
	}

	// RosePine.AddedBg = "#333c48" → 48;2;51;60;72
	if !strings.Contains(result, "48;2;51;60;72") {
		t.Fatal("expected RosePine added background")
	}

	if !strings.Contains(result, "   1 + ") {
		t.Fatalf("expected line to contain right-side format, got: %q", result)
	}
}
