package main

import (
	"fmt"
	"github.com/popstk/inke/internal/utils"
	"github.com/xanzy/go-gitlab"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func CompareGroup(git *gitlab.Client, conf Group, hostName string) {
	// check base group info
	utils.PrintInfo("check base group: %s", conf.BaseGroup)

	gp, _, err := git.Groups.GetGroup(conf.BaseGroup, nil)
	if err != nil {
		log.Fatalf("GetGroup %v err: %v", conf.BaseGroup, err)
	}

	utils.PrintInfo("GroupId: %d, Name: %s, Description: %s", gp.ID, gp.Name, gp.Description)

	ignores := make(map[string]struct{})
	for _, p := range conf.IgnoreDirs {
		full := strings.Join([]string{conf.BaseGroup, p}, "/")
		ignores[full] = struct{}{}
	}

	goPath := os.Getenv("GOPATH")
	base := filepath.Join(goPath, "src", hostName)

	// list all projects and check it
	everyThingOK := true
	err = ListGroupProjects(git, gp.ID, func(p *gitlab.Project) error {
		if _, ok := ignores[p.PathWithNamespace]; ok {
			return nil
		}

		full := []string{base}
		paths := strings.Split(p.PathWithNamespace, "/")
		for _, pp := range paths {
			full = append(full, pp)
		}

		realPath := strings.TrimPrefix(p.PathWithNamespace, conf.BaseGroup)
		if realPath[0] == '/' {
			realPath = realPath[1:]
		}

		local := filepath.Join(full...)
		if verbose {
			utils.PrintNormal(">[%d] %s => %s", p.ID, realPath, local)
		}

		_, err := os.Stat(local)
		if err == nil {
			return nil
		}

		if !os.IsNotExist(err) {
			return fmt.Errorf("os stat path %s err: %v", local, err)
		}

		everyThingOK = false

		if !clone {
			utils.PrintError("=> %s in %s ", realPath, local)
			return nil
		}

		args := []string{"git", "clone"}
		args = append(args, fmt.Sprintf("git@%s:%s.git", hostName, p.PathWithNamespace), local)
		utils.PrintNormal("exec| %s", strings.Join(args, " "))

		c := exec.Command(args[0], args[1:]...)
		output, err := c.CombinedOutput()
		if err != nil {
			return fmt.Errorf("exec err: %w", err)
		}

		utils.PrintWarn(string(output))
		return nil

	})
	if err != nil {
		log.Fatalf("Failed to ListGroupProjects: %v", err)
	}

	if everyThingOK {
		utils.PrintNormal("everything is ok!")
	}
}
