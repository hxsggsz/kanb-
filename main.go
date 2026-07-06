package main

import (
	"fmt"
	"os"

	tea "charm.land/bubbletea/v2"

	"kanba/cmd"
	_ "kanba/cmd/diff"
	_ "kanba/cmd/show"
)

func main() {
	if len(os.Getenv("DEBUG")) > 0 {
		if f, err := tea.LogToFile("debug.log", "debug"); err != nil {
			fmt.Fprintf(os.Stderr, "failed to open log file: %v\n", err)
			os.Exit(1)
		} else {
		defer f.Close()
		}
	}

	cmd.Execute()
}
