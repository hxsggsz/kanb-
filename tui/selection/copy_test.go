package selection

import "testing"

func TestCopyMsg_Content(t *testing.T) {
	msg := CopyMsg{Content: "hello world"}
	if msg.Content != "hello world" {
		t.Errorf("Content = %q, want %q", msg.Content, "hello world")
	}
}

func TestCopyMsg_EmptyContent(t *testing.T) {
	msg := CopyMsg{}
	if msg.Content != "" {
		t.Errorf("Content = %q, want empty string", msg.Content)
	}
}

func TestDelayedCopyCmd_ReturnsNonNil(t *testing.T) {
	cmd := DelayedCopyCmd("test")
	if cmd == nil {
		t.Error("DelayedCopyCmd returned nil")
	}
}
