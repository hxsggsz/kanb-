package git

import (
	"context"
	"errors"
	"fmt"
	"os/exec"
)

type Runner interface {
	Run(ctx context.Context, args ...string) (string, error)
}

type GitRunner struct {
	repoPath string
}

func NewGitRunner(repoPath string) *GitRunner {
	return &GitRunner{repoPath: repoPath}
}

func (r *GitRunner) Run(ctx context.Context, args ...string) (string, error) {
	cmd := exec.CommandContext(ctx, "git", args...)
	cmd.Dir = r.repoPath
	out, err := cmd.Output()
	if err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			return "", fmt.Errorf("git %s: %s: %w", args[0], string(exitErr.Stderr), exitErr)
		}
		return "", fmt.Errorf("git %s: %w", args[0], err)
	}
	return string(out), nil
}
