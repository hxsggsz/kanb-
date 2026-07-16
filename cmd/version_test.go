package cmd

import (
	"bytes"
	"strings"
	"testing"
)

func TestVersionCommandPrintsVersion(t *testing.T) {
	Version = "v9.9.9"
	defer func() { Version = "dev" }()

	buf := &bytes.Buffer{}
	RootCmd.SetOut(buf)
	RootCmd.SetArgs([]string{"version"})

	if err := RootCmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(buf.String(), "v9.9.9") {
		t.Errorf("expected output to contain version, got %q", buf.String())
	}
}
