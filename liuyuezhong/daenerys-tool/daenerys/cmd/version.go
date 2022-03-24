package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var (
	version = "v0.0.4"
)

func init() {
	rootCmd.AddCommand(versionCmd)
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number of daenerys tool",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("Daenerys Code Generator version=%s\n", version)
	},
}
