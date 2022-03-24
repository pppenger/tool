package cmd

import (
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(toolCmd)
	toolCmd.AddCommand(installCmd)
}

var toolCmd = &cobra.Command{
	Use:   "tool",
	Short: "Daenerys is a very fast static site generator",
	Long: `A Fast and Flexible Static Site Generator built with
love by spf13 and friends.
Complete documentation is available at http://hugo.spf13.com`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) < 1 {
			if err := cmd.Help(); err != nil {
				return err
			}
		}
		return nil
	},
}

var installCmd = &cobra.Command{
	Use:   "install",
	Short: "Daenerys is a very fast static site generator",
	Long: `A Fast and Flexible Static Site Generator built with
love by spf13 and friends.
Complete documentation is available at http://hugo.spf13.com`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) != 1 {
			if err := cmd.Help(); err != nil {
				return err
			}
		}
		return nil
	},
}
