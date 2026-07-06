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
	result := renderAlignedLine(fmtr, ln, 80, false, sh, "main.go")
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
	result := renderAlignedLine(fmtr, ln, 80, false, sh, "main.go")
	if !strings.Contains(result, "\x1b[48;5;22m") {
		t.Fatalf("expected green background (\\x1b[48;5;22m) for added line, got: %q", result)
	}
	if !strings.Contains(result, "\x1b[1m") && !strings.Contains(result, "\x1b[3") {
		t.Fatalf("expected syntax highlighting (\\x1b[1m or \\x1b[3), got: %q", result)
	}
}
