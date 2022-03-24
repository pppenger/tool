package main

import (
	"bytes"
	"golang.org/x/tools/go/vcs"
	"os/exec"
	"path/filepath"
	"strings"
)

type Vcs struct {
	vcs         *vcs.Cmd
	IdentifyCmd string
	DescribeCmd string
	DiffCmd     string
	ListCmd     string
	RootCmd     string
	ExistsCmd   string
	BranchCmd   string
	RemoteCmd   string
	CloneCmd    string
	ListTagsCmd string
}

var vcsGit = &Vcs{
	vcs:         vcs.ByCmd("git"),
	IdentifyCmd: "rev-parse HEAD",
	DescribeCmd: "describe --tags",
	//DiffCmd:     "diff {rev}",
	DiffCmd:     "diff",
	ListCmd:     "ls-files --full-name",
	RootCmd:     "rev-parse --show-cdup",
	ExistsCmd:   "cat-file -e {rev} .",
	BranchCmd:   "symbolic-ref --short HEAD ",
	RemoteCmd:   "remote set-url origin {addr}",
	CloneCmd:    "clone -b {tag} {repo}",
	ListTagsCmd: "tag",
}

func (v *Vcs) ClonebyTag(dir, tag, repo string) error {
	return v.run(dir, v.CloneCmd, "tag", tag, "repo", repo)
}

func (v *Vcs) ListTags(dir string) (string, error) {
	out, err := v.runOutput(dir, v.ListTagsCmd)
	return string(bytes.TrimSpace(out)), err
}

func (v *Vcs) Remote(dir, addr string) error {
	return v.run(dir, v.RemoteCmd, "addr", addr)
}

func (v *Vcs) Exists(dir, rev string) bool {
	err := v.runVerboseOnly(dir, v.ExistsCmd, "rev", rev)
	return err == nil
}

func absRoot(dir, out string) string {
	if filepath.IsAbs(out) {
		return filepath.Clean(out)
	}
	return filepath.Join(dir, out)
}

func (v *Vcs) Root(dir string) (string, error) {
	out, err := v.runOutput(dir, v.RootCmd)
	return absRoot(dir, string(bytes.TrimSpace(out))), err
}

func (v *Vcs) Describe(dir string) string {
	out, err := v.runOutputVerboseOnly(dir, v.DescribeCmd)
	if err != nil {
		return ""
	}
	return string(bytes.TrimSpace(out))
}

func (v *Vcs) IsDirty(dir string) bool {
	//out, err := v.runOutput(dir, v.DiffCmd, "rev", rev)
	out, err := v.runOutput(dir, v.DiffCmd)
	return err != nil || len(out) != 0
}

func (v *Vcs) Identify(dir string) (string, error) {
	out, err := v.runOutput(dir, v.IdentifyCmd)
	return string(bytes.TrimSpace(out)), err
}

func (v *Vcs) Branch(dir string) (string, error) {
	//	out, err := v.runOutput(dir, v.Branch)
	out, err := v.runOutputVerboseOnly(dir, v.BranchCmd)
	return string(bytes.TrimSpace(out)), err
}

// RevSync checks out the revision given by rev in dir.
// The dir must exist and rev must be a valid revision.
func (v *Vcs) RevSync(dir, rev string) error {
	return v.run(dir, v.vcs.TagSyncCmd, "tag", rev)
}

func (v *Vcs) RevDefaultSync(dir string) error {
	return v.run(dir, v.vcs.TagSyncDefault)
}

func (v *Vcs) Create(dir, repo string) error {
	return v.vcs.Create(dir, repo)
}

func (v *Vcs) CreateRev(dir, repo, rev string) error {
	return v.vcs.CreateAtRev(dir, repo, rev)
}

func (v *Vcs) Update(dir string) error {
	return v.vcs.Download(dir)
}

func (v *Vcs) TagSync(dir, tag string) error {
	return v.vcs.TagSync(dir, tag)
}

// run runs the command line cmd in the given directory.
// keyval is a list of key, value pairs.  run expands
// instances of {key} in cmd into value, but only after
// splitting cmd into individual arguments.
// If an error occurs, run prints the command line and the
// command's combined stdout+stderr to standard error.
// Otherwise run discards the command's output.
func (v *Vcs) run(dir string, cmdline string, kv ...string) error {
	_, err := v.run1(dir, cmdline, kv, true)
	return err
}

// runVerboseOnly is like run but only generates error output to standard error in verbose mode.
func (v *Vcs) runVerboseOnly(dir string, cmdline string, kv ...string) error {
	_, err := v.run1(dir, cmdline, kv, false)
	return err
}

// runOutput is like run but returns the output of the command.
func (v *Vcs) runOutput(dir string, cmdline string, kv ...string) ([]byte, error) {
	return v.run1(dir, cmdline, kv, true)
}

// runOutputVerboseOnly is like runOutput but only generates error output to standard error in verbose mode.
func (v *Vcs) runOutputVerboseOnly(dir string, cmdline string, kv ...string) ([]byte, error) {
	return v.run1(dir, cmdline, kv, false)
}

// run1 is the generalized implementation of run and runOutput.
func (v *Vcs) run1(dir string, cmdline string, kv []string, verbose bool) ([]byte, error) {
	m := make(map[string]string)
	for i := 0; i < len(kv); i += 2 {
		m[kv[i]] = kv[i+1]
	}
	args := strings.Fields(cmdline)
	for i, arg := range args {
		args[i] = expand(m, arg)
	}

	_, err := exec.LookPath(v.vcs.Cmd)
	if err != nil {
		doPrint("missing %s command.\n", v.vcs.Name)
		return nil, err
	}

	cmd := exec.Command(v.vcs.Cmd, args...)
	cmd.Dir = dir
	var buf bytes.Buffer
	cmd.Stdout = &buf
	cmd.Stderr = &buf
	err = cmd.Run()
	out := buf.Bytes()
	if err != nil {
		/*
			if verbose {
				fmt.Fprintf(os.Stderr, "# cd %s; %s %s\n", dir, v.vcs.Cmd, strings.Join(args, " "))
				os.Stderr.Write(out)
			}
		*/
		return nil, err
	}
	return out, nil
}

func expand(m map[string]string, s string) string {
	for k, v := range m {
		s = strings.Replace(s, "{"+k+"}", v, -1)
	}
	return s
}
