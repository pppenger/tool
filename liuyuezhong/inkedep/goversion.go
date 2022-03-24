package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// trimGoVersion and return the major version
func trimGoVersion(version string) (string, error) {
	if version == "devel" {
		return "devel", nil
	}
	if strings.HasPrefix(version, "devel+") || strings.HasPrefix(version, "devel-") {
		return strings.Replace(version, "devel+", "devel-", 1), nil
	}
	p := strings.Split(version, ".")
	if len(p) < 2 {
		return "", fmt.Errorf("Error determining major go version from: %q", version)
	}
	var split string
	switch {
	case strings.Contains(p[1], "beta"):
		split = "beta"
	case strings.Contains(p[1], "rc"):
		split = "rc"
	}
	if split != "" {
		p[1] = strings.Split(p[1], split)[0]
	}
	return p[0] + "." + p[1], nil
}

func getGoVersion() (string, error) {
	// Godep might have been compiled with a different
	// version, so we can't just use runtime.Version here.
	cmd := exec.Command("go", "version")
	cmd.Stderr = os.Stderr
	out, err := cmd.Output()
	return string(out), err
}

// goVersion returns the major version string of the Go compiler
// currently installed, e.g. "go1.5".
func GoVersion() (string, error) {
	out, err := getGoVersion()
	if err != nil {
		return "", err
	}
	gv := strings.Split(out, " ")
	if len(gv) < 4 {
		return "", fmt.Errorf("Error splitting output of `go version`: Expected 4 or more elements, but there are < 4: %q", out)
	}
	if gv[2] == "devel" {
		return trimGoVersion(gv[2] + gv[3])
	}
	return trimGoVersion(gv[2])
}
