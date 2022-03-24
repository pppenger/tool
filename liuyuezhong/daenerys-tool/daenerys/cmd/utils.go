package cmd

import (
	"os"
	"os/exec"
	"strings"
)

func install(exe, dir string, path map[string]string, args ...string) error {
	cmd := exec.Command(exe)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Dir = dir
	cmd.Args = append([]string{exe}, args...)
	cmd.Env = os.Environ()
	for k, v := range path {
		cmd.Env = append(cmd.Env, strings.Join([]string{k, v}, "="))
	}
	if err := cmd.Run(); err != nil {
		return err
	}
	return nil
}

func isInstall(exe string) bool {
	_, err := exec.LookPath(exe)
	if err != nil {
		return false
	}
	return true
}
