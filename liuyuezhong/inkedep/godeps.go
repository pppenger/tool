package main

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"
)

type CustomConfig struct {
	SpecificTag string `json:"SpecificTag"`
	SpecificRev string `json:"SpecificRev"`
}

type Dependence struct {
	ImportPath string
	Comment    string       `json:",omitempty"`
	Rev        string       `json:",omitempty"` // VCS-specific commit ID.
	Tag        string       `json:",omitempty"`
	Remote     string       `json:",omitempty"`
	Custom     CustomConfig `json:"Custom"`
}

func (d *Dependence) changePath() string {
	for _, repo := range rConfig.RepoExchanged {
		if strings.HasPrefix(d.ImportPath, repo.Old) {
			if len(repo.Branch) > 0 {
				d.Custom.SpecificTag = repo.Branch
			}
			return strings.Replace(d.ImportPath, repo.Old, repo.New, 1)
		}
	}
	// default
	if strings.HasPrefix(d.ImportPath, "golang.org/x") {
		return strings.Replace(d.ImportPath, "golang.org/x", "github.com/golang", 1)
	}

	return d.ImportPath
}

func (d *Dependence) Create(dir string) error {
	rev := d.Custom.SpecificTag
	if len(d.Custom.SpecificRev) > 0 {
		rev = d.Custom.SpecificRev
	} else if len(d.Rev) > 0 { // 兼容v1
		rev = d.Rev
	}
	importpath := d.changePath()
	var repo string
	var err error
	if d.Remote != "" {
		repo = d.Remote
	} else {
		repo, _, err = FindRepoCacheRoot(importpath)
	}
	if err == nil && len(repo) > 0 {
		if err := vcsGit.CreateRev(dir, repo, rev); err == nil {
			doPrint(">> on cache, real repo %s, into dst %s on GOPATH.", importpath, d.ImportPath)
			return nil
		}
		Remove(dir)
	}

	if ok, _ := IsPathExist(dir); ok {
		return nil
	}

	// remote repo
	var remoteRepo string
	var e error
	if remoteRepo, _, e = FindRepoRoot(importpath); e != nil {
		doError("find repo root fail, %q", e.Error())
		return e
	}
	doPrint(">> from remote, real repo %s, into dst %s on GOPATH.", remoteRepo, d.ImportPath)
	if e := vcsGit.CreateRev(dir, remoteRepo, rev); e != nil {
		if strings.Contains(e.Error(), "already exists") {
			return nil
		}
		return fmt.Errorf("download or checkout fail, repo %s, rev %s", remoteRepo, rev)
	}

	return nil
}

type Godeps struct {
	ImportPath     string        // project path
	GoVersion      string        // go version
	InkedepVersion string        // inkedep version
	Deps           []*Dependence // dependence's path
}

func (g *Godeps) SaveGodeps(path string) error {
	sort.Slice(g.Deps, func(i, j int) bool {
		return strings.Compare(g.Deps[i].ImportPath, g.Deps[j].ImportPath) > 0
	})
	for _, dep := range g.Deps {
		doPrint("save dep, ImportPath:%s\n", dep.ImportPath)
	}
	fd, err := os.Create(path)
	if err != nil {
		return err
	}
	b, _ := json.MarshalIndent(g, "", "  ")
	_, err = fd.Write(b)
	return err
}

func (g *Godeps) OpenGodeps(path string) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()
	dec := json.NewDecoder(file)
	return dec.Decode(g)
}
