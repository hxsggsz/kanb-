package diff

import (
	"kanba/cmd"

	"github.com/spf13/cobra"
)

var DiffCmd = &cobra.Command{
	Use:   "diff [flags]",
	Short: "Show a git diff",
	Long:  `Show a git diff. Passes flags directly through to git diff.`,
	Run: func(_ *cobra.Command, args []string) {
		cmd.RunTUI(args)
	},
	DisableFlagParsing: true,
}

func init() {
	cmd.RootCmd.AddCommand(DiffCmd)
}
