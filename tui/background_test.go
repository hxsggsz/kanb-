package tui

import (
	"strings"
	"testing"

	"charm.land/lipgloss/v2"
	"kanba/git"
)

func TestRenderStyledLineAddsBackground(t *testing.T) {
	result := renderStyledLine("", "hello", 20, git.KindAdded, false, false, nil, "", RosePine)
	// RosePine.AddedBg is "#333c48" → 48;2;51;60;72
	if !strings.Contains(result, "48;2;51;60;72") {
		t.Fatalf("expected RosePine added background, got: %q", result)
	}
	if lipgloss.Width(result) != 20 {
		t.Fatalf("expected visible width 20, got %d", lipgloss.Width(result))
	}
}

func TestRenderStyledLineNoExtraWidth(t *testing.T) {
	result := renderStyledLine("", "hello world", 5, git.KindContext, false, false, nil, "", RosePine)
	if lipgloss.Width(result) != 11 {
		t.Fatalf("expected visible width 11 (no padding needed), got %d", lipgloss.Width(result))
	}
}

func TestRenderStyledLineCursorHighlight(t *testing.T) {
	result := renderStyledLine("", "hello", 10, git.KindContext, false, true, nil, "", RosePine)
	// Cursor bg on context = CursorBgFor(PanelBg) = blendHex("#403d52", "#1f1d2e", 0.75) → 48;2;56;53;73
	if !strings.Contains(result, "48;2;56;53;73") {
		t.Fatalf("expected cursor background blended over panel, got: %q", result)
	}
	if lipgloss.Width(result) != 10 {
		t.Fatalf("expected visible width 10, got %d", lipgloss.Width(result))
	}
}

func TestThemeBgFor(t *testing.T) {
	tests := []struct {
		kind   git.LineKind
		isLeft bool
		expect string
	}{
		{git.KindAdded, true, ""},
		{git.KindAdded, false, "#333c48"},
		{git.KindDeleted, true, "#312e3f"},
		{git.KindDeleted, false, ""},
		{git.KindModified, true, "#312e3f"},
		{git.KindModified, false, "#333c48"},
		{git.KindContext, true, ""},
		{git.KindContext, false, ""},
	}
	for _, tt := range tests {
		got := RosePine.BgFor(tt.kind, tt.isLeft)
		if got != tt.expect {
			t.Errorf("RosePine.BgFor(%v, %v) = %q, want %q", tt.kind, tt.isLeft, got, tt.expect)
		}
	}
}
