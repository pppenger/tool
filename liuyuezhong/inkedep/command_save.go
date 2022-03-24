package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

var cmdSave = &Command{
	Run:          runSave,
	Name:         "save",
	Short:        "该命令在项目代码入口函数(main)所在路径下执行,save会分析项目的依赖,并把依赖信息保存到Godeps/Godeps.json文件中.",
	OnlyInGOPATH: true,
}

func runSave(args []string, ctx *Context) error {
	if ok, _ := IsPathExist("vendor"); ok {
		doWarning("vendor directory exist, please remove it first")
		return fmt.Errorf("vendor directory exist")
	}

	goDeps := &Godeps{}
	doPrint("Getting go version ...")
	cv, err := GoVersion()
	if err != nil {
		return err
	}
	goDeps.GoVersion = cv
	doPrint("Getting go version done")

	doPrint("Getting project git root, workdir:%s", ctx.WorkDir)
	root, err := FindGitRoot(ctx.WorkDir)
	if err != nil {
		return err
	}
	doPrint("Getting git root done, root:%s", root)

	path := filepath.Join(ctx.GOPATH, "src")
	project, err := filepath.Rel(path, root)
	if err != nil {
		return err
	}

	// 兼容windows
	project = filepath.ToSlash(project)

	doPrint("Local project import path: %s", project)
	goDeps.ImportPath = project
	goDeps.InkedepVersion = "2.0"

	// 之前的配置
	loadOldDepConfig()

	start := time.Now()
	doPrint("Scan all deps under project root: %s", root)
	deps := ctx.loadDeps(root, project, false)
	doPrint("Scan all deps done, cost %.2fs", time.Since(start).Seconds())
	for _, v := range deps {
		vv := v
		goDeps.Deps = append(goDeps.Deps, vv)
	}
	if err := os.Mkdir("Godeps", 0755); err != nil && !os.IsExist(err) {
		return err
	}
	if err := goDeps.SaveGodeps("Godeps/Godeps.json"); err != nil {
		return err
	}
	return nil
}

type oldConfig struct {
	tag string
	rev string
}

var userDefinedConfig map[string]oldConfig

func loadOldDepConfig() {
	userDefinedConfig = make(map[string]oldConfig)
	filePath := "./Godeps/Godeps.json"
	g := &Godeps{}
	if err := g.OpenGodeps(filePath); err != nil {
		return
	}
	for _, d := range g.Deps {
		s := oldConfig{}
		if len(d.Custom.SpecificTag) > 0 {
			s.tag = d.Custom.SpecificTag
		}
		if len(d.Custom.SpecificRev) > 0 {
			s.rev = d.Custom.SpecificRev
		}
		userDefinedConfig[d.ImportPath] = s
	}
}

func saveDepObject(depPath string, ctx *Context) (*Dependence, error) {
	vcs := vcsGit
	dir := ctx.GopathSrc(depPath)
	//if vcs.IsDirty(dir) {
	//	doWarning("\033[43;34m dirty working tree: %s \033[0m", dir)
	//}

	branch, _ := vcs.Branch(dir)
	//if len(branch) > 0 && branch != "master" {
	//	doWarning("\033[43;34m module %s not on 'master', using branch %s \033[0m", depPath, branch)
	//}

	cc := CustomConfig{SpecificTag: userDefinedConfig[depPath].tag, SpecificRev: userDefinedConfig[depPath].rev}

	if len(cc.SpecificTag) == 0 && len(cc.SpecificRev) == 0 {
		cc.SpecificTag = "master"
		if len(branch) > 0 {
			cc.SpecificTag = branch
		}
	}

	if strings.Contains(depPath, "github.com/go-playground/validator") {
		cc.SpecificTag = "v9"
	}
	if strings.Contains(depPath, "github.com/go-check/check") {
		cc.SpecificTag = "v1"
	}
	if strings.Contains(depPath, "github.com/apache/thrift") {
		cc.SpecificTag = "0.10.0"
	}
	if strings.Contains(depPath, "git.apache.org/thrift.git") {
		cc.SpecificTag = "0.10.0"
	}

	return &Dependence{
		ImportPath: depPath,
		Custom:     cc,
	}, nil
}
