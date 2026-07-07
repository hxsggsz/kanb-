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
	result := 		renderAlignedLine(fmtr, ln, 80, false, sh, "main.go", 0, false)
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
	result := 		renderAlignedLine(fmtr, ln, 80, false, sh, "main.go", 0, false)

	rightBG := "\x1b[48;5;22m"
	if !strings.Contains(result, rightBG) {
		t.Fatalf("expected green background (%s) for added line, got: %q", rightBG, result)
	}

	sep := " │ "
	idx := strings.Index(result, sep)
	if idx < 0 {
		t.Fatalf("expected separator %q in result: %q", sep, result)
	}
	rightSide := result[idx+len(sep):]
	if !strings.HasPrefix(rightSide, rightBG) {
		t.Fatalf("expected right side to start with green background %s, got: %q", rightBG, rightSide[:20])
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
	result := 		renderAlignedLine(fmtr, ln, 80, false, sh, "main.go", 0, false)

	leftBG := "\x1b[48;5;52m"
	if !strings.Contains(result, leftBG) {
		t.Fatalf("expected red background (%s) for deleted line, got: %q", leftBG, result)
	}

	if !strings.HasPrefix(result, leftBG) {
		t.Fatalf("expected result to start with red background %s, got: %q", leftBG, result[:20])
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
	result := renderAlignedLine(fmtr, ln, 80, false, sh, "main.go", 0, true)

	if strings.Contains(result, " │ ") {
		t.Fatal("single-panel result should not contain separator")
	}

	if !strings.Contains(result, "\x1b[48;5;22m") {
		t.Fatal("expected green background for added line")
	}

	if !strings.Contains(result, "   1 + ") {
		t.Fatalf("expected line to contain right-side format, got: %q", result)
	}
}
