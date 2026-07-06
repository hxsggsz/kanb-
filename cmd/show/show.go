package show

import (
	"kanba/cmd"

	"github.com/spf13/cobra"
)

var ShowCmd = &cobra.Command{
	Use:   "show [ref]",
	Short: "Show a commit",
	Long:  `Show the diff for a specific commit or reference.`,
	Run: func(_ *cobra.Command, args []string) {
		cmd.RunTUI(append([]string{"show"}, args...))
	},
}

func init() {
	cmd.RootCmd.AddCommand(ShowCmd)
}
