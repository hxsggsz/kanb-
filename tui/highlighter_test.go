package tui

import (
	"strings"
	"testing"
)

func TestHighlighterHighlightsGoCode(t *testing.T) {
	sh := NewSyntaxHighlighter()
	code := "func main() {}"
	result := sh.Highlight(code, "main.go")
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
	result := sh.Highlight(code, "unknown.xyz")
	if result != code {
		t.Fatal("expected plain code when no lexer matches")
	}
}

func TestHighlighterCachesLexers(t *testing.T) {
	sh := NewSyntaxHighlighter()
	r1 := sh.Highlight("func a() {}", "a.go")
	r2 := sh.Highlight("func b() {}", "b.go")
	if !strings.Contains(r1, "\x1b[") {
		t.Fatal("expected ANSI in first result")
	}
	if !strings.Contains(r2, "\x1b[") {
		t.Fatal("expected ANSI in second result")
	}
	r3 := sh.Highlight("func c() {}", "a.go")
	if r3 == "" {
		t.Fatal("expected non-empty cached result")
	}
	if !strings.Contains(r3, "\x1b[") {
		t.Fatal("expected ANSI in cached result")
	}
}

func TestHighlighterEmptyCode(t *testing.T) {
	sh := NewSyntaxHighlighter()
	result := sh.Highlight("", "main.go")
	if result != "" {
		t.Fatal("expected empty string for empty input")
	}
}
