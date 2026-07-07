package tui

import (
	"strings"
	"testing"

	"charm.land/lipgloss/v2"
)

func TestHighlighterHighlightsGoCode(t *testing.T) {
	sh := NewSyntaxHighlighter()
	code := "func main() {}"
	result := sh.HighlightWithStyle(code, "main.go", lipgloss.NewStyle())
	if result == code {
		t.Fatal("expected ANSI escape sequences in highlighted Go code")
	}
	if !strings.Contains(result, "\x1b[") {
		t.Fatal("highlighted output should contain ANSI escape codes")
	}
}

func TestHighlighterFallbackForUnknownExtension(t *testing.T) {
	sh := NewSyntaxHighlighter()
	code := "some random text without a known extension"
	result := sh.HighlightWithStyle(code, "unknown.xyz", lipgloss.NewStyle())
	if !strings.HasPrefix(result, "\x1b[") && result != code {
		t.Fatal("expected plain code or styled output when no lexer matches")
	}
}

func TestHighlighterCachesLexers(t *testing.T) {
	sh := NewSyntaxHighlighter()
	r1 := sh.HighlightWithStyle("func a() {}", "a.go", lipgloss.NewStyle())
	r2 := sh.HighlightWithStyle("func b() {}", "b.go", lipgloss.NewStyle())
	if !strings.Contains(r1, "\x1b[") {
		t.Fatal("expected ANSI in first result")
	}
	if !strings.Contains(r2, "\x1b[") {
		t.Fatal("expected ANSI in second result")
	}
	r3 := sh.HighlightWithStyle("func c() {}", "a.go", lipgloss.NewStyle())
	if r3 == "" {
		t.Fatal("expected non-empty cached result")
	}
	if !strings.Contains(r3, "\x1b[") {
		t.Fatal("expected ANSI in cached result")
	}
}

func TestHighlighterEmptyCode(t *testing.T) {
	sh := NewSyntaxHighlighter()
	result := sh.HighlightWithStyle("", "main.go", lipgloss.NewStyle())
	if result != "" {
		t.Fatal("expected empty string for empty input")
	}
}
