package diff

import (
	"strings"
	"testing"

	models "kanba/tui/models"

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
	fmtr := DefaultFormatters[ln.Kind]
	result := RenderAlignedLine(fmtr, ln, 80, sh, "main.go", 0, false, false, models.GetTheme("rose-pine"))
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
	fmtr := DefaultFormatters[ln.Kind]
	result := RenderAlignedLine(fmtr, ln, 80, sh, "main.go", 0, false, false, models.GetTheme("rose-pine"))

	if !strings.Contains(result, "48;2;156;207;216") {
		t.Fatalf("expected added background, got: %q", result)
	}

	rightPrefix := "\x1b[38;2;156;207;216;48;2;156;207;216m"
	if !strings.Contains(result, rightPrefix) {
		t.Fatalf("expected right panel to contain %s, got: %q", rightPrefix, result)
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
	fmtr := DefaultFormatters[ln.Kind]
	result := RenderAlignedLine(fmtr, ln, 80, sh, "main.go", 0, false, false, models.GetTheme("rose-pine"))

	if !strings.Contains(result, "48;2;235;111;146") {
		t.Fatalf("expected removed background, got: %q", result)
	}

	leftPrefix := "\x1b[38;2;235;111;146;48;2;235;111;146m"
	if !strings.Contains(result, leftPrefix) {
		t.Fatalf("expected result to contain %s, got: %q", leftPrefix, result)
	}
}

func TestRenderAlignedLineSinglePanel(t *testing.T) {
	sh := NewSyntaxHighlighter()
	ln := git.AlignedLine{
		NewLineNum: 1,
		NewContent: "func main() {",
		Kind:       git.KindAdded,
	}
	fmtr := DefaultFormatters[ln.Kind]
	result := RenderAlignedLine(fmtr, ln, 80, sh, "main.go", 0, true, false, models.GetTheme("rose-pine"))

	if !strings.Contains(result, "48;2;156;207;216") {
		t.Fatal("expected added background")
	}

	if !strings.Contains(result, "   1 + ") {
		t.Fatalf("expected line to contain right-side format, got: %q", result)
	}
}
