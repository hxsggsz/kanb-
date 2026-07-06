package git

import (
	"fmt"
	"os/exec"
	"strings"
)

func Diff(repoPath string, args []string) ([]FileDiff, error) {
	var cmdArgs []string
	if len(args) > 0 && args[0] == "show" {
		cmdArgs = append([]string{"show", "--no-color", "--unified=3"}, args[1:]...)
	} else {
		cmdArgs = append([]string{"diff", "--no-color", "--unified=3"}, args...)
	}

	cmd := exec.Command("git", cmdArgs...)
	cmd.Dir = repoPath
	out, err := cmd.Output()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return nil, fmt.Errorf("git %s failed: %s: %w", cmdArgs[0], string(exitErr.Stderr), exitErr)
		}
		return nil, fmt.Errorf("git %s: %w", cmdArgs[0], err)
	}

	raw := string(out)
	if raw == "" {
		return []FileDiff{}, nil
	}

	return parseRawDiff(strings.Split(raw, "\n"))
}

func RepoRoot(path string) (string, error) {
	cmd := exec.Command("git", "rev-parse", "--show-toplevel")
	cmd.Dir = path
	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("not a git repository: %w", err)
	}
	return strings.TrimSpace(string(out)), nil
}
