package git

import "strings"

type RepoRootCommand struct{}

func (c *RepoRootCommand) Args() []string {
	return []string{"rev-parse", "--show-toplevel"}
}

func (c *RepoRootCommand) Parse(raw string) (string, error) {
	return strings.TrimSpace(raw), nil
}
