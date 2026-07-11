package selection

import "testing"

func TestDelayedCopyCmd_ReturnsNonNil(t *testing.T) {
	cmd := DelayedCopyCmd()
	if cmd == nil {
		t.Error("DelayedCopyCmd returned nil")
	}
}
