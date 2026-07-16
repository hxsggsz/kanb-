package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// Version is the CLI version, injected at build time via:
// -ldflags "-X kanba/cmd.Version=vX.Y.Z"
var Version = "dev"

var VersionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the kanba version",
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Fprintln(cmd.OutOrStdout(), Version)
		return nil
	},
}

func init() {
	RootCmd.Version = Version
	RootCmd.AddCommand(VersionCmd)
}
