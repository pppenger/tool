package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(updateCmd)
}

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "update daenerys tool.",
	Args: func(cmd *cobra.Command, args []string) (err error) {
		// no args
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		defer func() {
			fmt.Printf("\n\nDaenerys Code Generator latest version=%s\n\n", version)
		}()
		return selfUpdate()
	},
}

func selfUpdate() error {
	var p string
	if gobin, exist := os.LookupEnv("GOBIN"); !exist {
		p = filepath.Join(os.Getenv("GOPATH"), "bin")
	} else {
		p = gobin
	}
	return install("go", "/tmp", map[string]string{"GOBIN": p}, "get", "-v", "-u", "git.inke.cn/BackendPlatform/daenerys-tool/...")
}
