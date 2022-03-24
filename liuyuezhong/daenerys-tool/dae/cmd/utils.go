package cmd

import (
	"os"
	"os/exec"
)

func run(command string, args []string) error {
	cmd := exec.Command(command)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Args = append([]string{command}, args...)
	cmd.Env = os.Environ()
	return cmd.Run()
}
