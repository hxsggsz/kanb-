package cmd

import (
	"context"
	"fmt"
	"os"

	tea "charm.land/bubbletea/v2"
	"github.com/spf13/cobra"

	"kanba/config"
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

	runner := git.NewGitRunner(cwd)
	repoPath, err := git.Execute(context.Background(), runner, &git.RepoRootCommand{})
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "warning: config error: %v\n", err)
	}

	p := tea.NewProgram(tui.New(repoPath, gitArgs, cfg, Version))
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
