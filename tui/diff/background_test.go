package diff

import (
	"strings"
	"testing"

	models "kanba/tui/models"

	"charm.land/lipgloss/v2"
	"kanba/git"
)

func TestRenderStyledLineAddsBackground(t *testing.T) {
	result := renderStyledLine("", "hello", 20, git.KindAdded, false, false, nil, "", models.GetTheme("rose-pine"))
	if !strings.Contains(result, "48;2;51;60;72") {
		t.Fatalf("expected added background, got: %q", result)
	}
	if lipgloss.Width(result) != 20 {
		t.Fatalf("expected visible width 20, got %d", lipgloss.Width(result))
	}
}

func TestRenderStyledLineNoExtraWidth(t *testing.T) {
	result := renderStyledLine("", "hello world", 5, git.KindContext, false, false, nil, "", models.GetTheme("rose-pine"))
	if lipgloss.Width(result) != 11 {
		t.Fatalf("expected visible width 11 (no padding needed), got %d", lipgloss.Width(result))
	}
}

func TestRenderStyledLineCursorHighlight(t *testing.T) {
	result := renderStyledLine("", "hello", 10, git.KindContext, false, true, nil, "", models.GetTheme("rose-pine"))
	if !strings.Contains(result, "48;2;56;53;73") {
		t.Fatalf("expected cursor background blended over panel, got: %q", result)
	}
	if lipgloss.Width(result) != 10 {
		t.Fatalf("expected visible width 10, got %d", lipgloss.Width(result))
	}
}

func TestThemeBgFor(t *testing.T) {
	theme := models.GetTheme("rose-pine")
	tests := []struct {
		kind   git.LineKind
		isLeft bool
		expect string
	}{
		{git.KindAdded, true, ""},
		{git.KindAdded, false, "#333c48"},
		{git.KindDeleted, true, "#43293a"},
		{git.KindDeleted, false, ""},
		{git.KindModified, true, "#43293a"},
		{git.KindModified, false, "#333c48"},
		{git.KindContext, true, ""},
		{git.KindContext, false, ""},
	}
	for _, tt := range tests {
		got := theme.BgFor(tt.kind, tt.isLeft)
		if got != tt.expect {
			t.Errorf("BgFor(%v, %v) = %q, want %q", tt.kind, tt.isLeft, got, tt.expect)
		}
	}
}
