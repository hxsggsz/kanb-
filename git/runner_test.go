package git

import (
	"context"
	"errors"
	"testing"
)

type MockRunner struct {
	Output string
	Err    error
}

func (m *MockRunner) Run(_ context.Context, args ...string) (string, error) {
	return m.Output, m.Err
}

type testCommand struct {
	args  []string
	parse func(string) (string, error)
}

func (c *testCommand) Args() []string                        { return c.args }
func (c *testCommand) Parse(out string) (string, error) { return c.parse(out) }

func TestExecuteSuccess(t *testing.T) {
	runner := &MockRunner{Output: "hello"}
	cmd := &testCommand{args: []string{"echo", "hello"}, parse: func(s string) (string, error) {
		return s, nil
	}}
	result, err := Execute(context.Background(), runner, cmd)
	if err != nil {
		t.Fatal(err)
	}
	if result != "hello" {
		t.Errorf("expected 'hello', got %q", result)
	}
}

func TestExecuteRunnerError(t *testing.T) {
	runner := &MockRunner{Err: errors.New("boom")}
	cmd := &testCommand{parse: func(s string) (string, error) { return s, nil }}
	_, err := Execute(context.Background(), runner, cmd)
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestExecuteParseError(t *testing.T) {
	runner := &MockRunner{Output: "bad data"}
	cmd := &testCommand{parse: func(_ string) (string, error) {
		return "", errors.New("parse error")
	}}
	_, err := Execute(context.Background(), runner, cmd)
	if err == nil || err.Error() != "parse error" {
		t.Fatalf("expected 'parse error', got %v", err)
	}
}
