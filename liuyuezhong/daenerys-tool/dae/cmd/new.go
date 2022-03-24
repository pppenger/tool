package cmd

import (
	"errors"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

var (
	Proto string
	Dir   string
	Type  string
	Name  string
)

var newCmd = &cobra.Command{
	Use:   "new",
	Short: "New project",
	Long:  `New project`,
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) != 1 {
			return errors.New("requires a project name")
		}
		Name = args[0]
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		var arg = []string{"new", Name}
		cmd.Flags().Visit(func(flag *pflag.Flag) {
			f := fmt.Sprintf("--%s=%s", flag.Name, flag.Value)
			arg = append(arg, f)
		})
		err := run("daenerys", arg)
		if err != nil {
			panic(fmt.Sprintf("New project error,err(%v)", err))
		}

	},
}

func init() {
	pwd, _ := os.Getwd()
	newCmd.Flags().StringVar(&Proto, "proto", "", "whether to use protobuf to create rpc project")
	newCmd.Flags().StringVar(&Dir, "project-dir", pwd, "project directory to create project")
	newCmd.Flags().StringVar(&Type, "type", "http", "http or rpc, default http")
	rootCmd.AddCommand(newCmd)

}
