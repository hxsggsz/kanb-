package main

import (
	"fmt"
	"os"

	tea "charm.land/bubbletea/v2"

	"kanba/git"
	"kanba/tui"
)

func main() {
	if len(os.Getenv("DEBUG")) > 0 {
		f, err := tea.LogToFile("debug.log", "debug")
		if err != nil {
			fmt.Fprintf(os.Stderr, "failed to open log file: %v\n", err)
			os.Exit(1)
		}
		defer f.Close()
	}

	args := os.Args[1:]
	gitArgs := []string{}

	if len(args) == 0 || args[0] == "diff" {
		gitArgs = args
	} else if args[0] == "show" {
		gitArgs = args
	} else if args[0] == "--help" || args[0] == "-h" {
		printUsage()
		return
	} else {
		printUsage()
		return
	}

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

func printUsage() {
	fmt.Println(`kanba — TUI git diff viewer

Usage:
  kanba              Show unstaged changes
  kanba diff         Show unstaged changes
  kanba diff --staged  Show staged changes
  kanba show [ref]   Show a commit
  kanba --help       Show this help`)
}
