package tui

import (
	"strings"
	"testing"

	"charm.land/lipgloss/v2"
	"kanba/git"
)

func TestInjectBackgroundReplacesReset(t *testing.T) {
	input := "\x1b[33mfunc\x1b[0m \x1b[34mmain\x1b[0m()"
	result := injectBackground(input, "42")
	if !strings.Contains(result, "\x1b[0m\x1b[42m") {
		t.Fatalf("expected \\x1b[0m\\x1b[42m in result, got: %q", result)
	}
}

func TestInjectBackgroundEmptyCode(t *testing.T) {
	input := "hello world"
	result := injectBackground(input, "")
	if result != input {
		t.Fatal("empty bgCode should return unchanged")
	}
}

func TestInjectCursorWrapsWithReverse(t *testing.T) {
	input := "hello"
	result := injectCursor(input)
	if !strings.HasPrefix(result, "\x1b[7m") {
		t.Fatalf("expected \\x1b[7m prefix, got: %q", result)
	}
	if !strings.HasSuffix(result, "\x1b[0m") {
		t.Fatalf("expected \\x1b[0m suffix, got: %q", result)
	}
}

func TestPadToWidthAddsBackground(t *testing.T) {
	input := "hello"
	result := padToWidth(input, 20, "42")
	if !strings.HasSuffix(result, "\x1b[0m") {
		t.Fatalf("expected \\x1b[0m suffix, got: %q", result)
	}
	if !strings.Contains(result, "\x1b[42m") {
		t.Fatalf("expected \\x1b[42m for padding, got: %q", result)
	}
	if lipgloss.Width(result) != 20 {
		t.Fatalf("expected visible width 20, got %d", lipgloss.Width(result))
	}
}

func TestPadToWidthNoPaddingNeeded(t *testing.T) {
	input := "hello world"
	result := padToWidth(input, 5, "42")
	if !strings.HasSuffix(result, "\x1b[0m") {
		t.Fatalf("expected \\x1b[0m suffix even when no padding, got: %q", result)
	}
}

func TestBackgroundFor(t *testing.T) {
	tests := []struct {
		kind   git.LineKind
		isLeft bool
		expect string
	}{
		{git.KindAdded, true, ""},
		{git.KindAdded, false, "48;5;22"},
		{git.KindDeleted, true, "48;5;52"},
		{git.KindDeleted, false, ""},
		{git.KindModified, true, "48;5;52"},
		{git.KindModified, false, "48;5;22"},
		{git.KindContext, true, ""},
		{git.KindContext, false, ""},
	}
	for _, tt := range tests {
		got := backgroundFor(tt.kind, tt.isLeft)
		if got != tt.expect {
			t.Errorf("backgroundFor(%v, %v) = %q, want %q", tt.kind, tt.isLeft, got, tt.expect)
		}
	}
}
