package cmd

import (
	"fmt"
	"os"

	tea "charm.land/bubbletea/v2"
	"github.com/spf13/cobra"

	"kanba/git"
	"kanba/tui"
)

var RootCmd = &cobra.Command{
	Use:   "kanba",
	Short: "TUI git diff viewer",
	Long:  `kanba is a terminal UI for browsing git diffs.`,
	Run: func(cmd *cobra.Command, args []string) {
		RunTUI([]string{})
	},
}

func RunTUI(gitArgs []string) {
	cwd, err := os.Getwd()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error getting current directory: %v\n", err)
		os.Exit(1)
	}

	repoPath, err := git.RepoRoot(cwd)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	p := tea.NewProgram(tui.New(repoPath, gitArgs))
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

func Execute() {
	if err := RootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
