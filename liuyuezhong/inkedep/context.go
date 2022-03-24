package main

import (
	"errors"
	"fmt"
	"go/build"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"
)

type Context struct {
	WorkDir string
	GOPATH  string
	GOROOT  string
	Out     io.Writer
	cache   map[string]struct{}
}

func NewCtx() (*Context, error) {
	ctx := &Context{}
	ctx.Out = os.Stdout
	if err := ctx.setPaths(); err != nil {
		return ctx, err
	}
	ctx.cache = map[string]struct{}{}
	return ctx, nil
}

func WalkDir(parentPath string) (subDirs []string, err error) {
	start := time.Now()
	subDirs = make([]string, 0, 100)
	err = filepath.Walk(parentPath, func(absName string, fi os.FileInfo, err error) error { // 遍历目录
		if strings.Contains(absName, ".git") ||
			strings.Contains(absName, "vendor") ||
			strings.Contains(absName, "Godeps") {
			return nil
		}
		if fi.IsDir() { // 子目录
			subDirs = append(subDirs, absName)
		}
		return nil
	})
	doPrint("walking on sub dir: %s, find: %d, cost: %v", parentPath, len(subDirs), time.Since(start).String())
	return subDirs, err
}

func (c *Context) ScanDeps(runPath, projectModule string, reScan bool) ([]string, error) {
	start := time.Now()
	err := c.deps(runPath, projectModule)
	if err != nil {
		if strings.Contains(err.Error(), "no buildable Go source files") {
			// 检查子目录
			subDir, _ := WalkDir(c.WorkDir)
			for _, dir := range subDir {
				dd := dir
				c.deps(dd, projectModule)
			}
		} else {
			return nil, err
		}
	}
	doPrint("recursive scan deps done, cost: %s, path: %s", time.Since(start).String(), runPath)

	c.simplify(projectModule)

	result := map[string]bool{}
	for p := range base1 {
		result[p] = true
	}
	for p := range base2 {
		result[p] = true
	}
	// 基础库1
	if baseLib(projectModule, baseModulePath) || reScan {
		for p := range base1deps {
			if _, ok := others[p]; ok {
				delete(others, p)
			}
			result[p] = true
		}

		// 基础库2
	} else if baseLib(projectModule, baseModulePath2) || reScan {
		for p := range base2deps {
			if _, ok := base1deps[p]; ok {
				delete(others, p)
				continue
			}
			result[p] = true
		}

	} else {
		for k := range others {
			if _, ok := base1deps[k]; ok {
				delete(others, k)
			}
			if _, ok := base2deps[k]; ok {
				delete(others, k)
			}
			result[k] = true
		}
	}

	for p := range others {
		result[p] = true
	}

	bar := make([]string, 0, 100)
	for m := range result {
		bar = append(bar, m)
	}
	return bar, nil
}

var (
	base1     = map[string]bool{}
	base1deps = map[string]bool{}
	base2     = map[string]bool{}
	base2deps = map[string]bool{}
	others    = map[string]bool{}
)

var lock = sync.Mutex{}
var hadScaned = sync.Map{}

func (c *Context) deps(dir, module string) error {
	if _, ok := hadScaned.Load(dir); ok {
		return nil
	}

	if ok, _ := c.buildInPackage(module); ok {
		return nil
	}
	// 在dir下扫描依赖,忽略vendor
	pkg, err := build.ImportDir(dir, build.IgnoreVendor)
	if err != nil {
		return err
	}

	for _, p := range pkg.Imports {
		if p == "C" || build.IsLocalImport(p) {
			continue
		}
		if ok, _ := c.buildInPackage(p); ok {
			continue
		}
		if strings.HasPrefix(p, module) {
			goto CON
		}
		// 基础库所依赖的三方包
		if baseLib(module, baseModulePath) && !baseLib(p, baseModulePath) {
			lock.Lock()
			base1deps[p] = true
			lock.Unlock()
		} else if baseLib(module, baseModulePath2) && !baseLib(p, baseModulePath2) {
			lock.Lock()
			base2deps[p] = true
			lock.Unlock()
		} else {
			lock.Lock()
			others[p] = true
			lock.Unlock()
		}
	CON:
		subDir := c.GopathSrc(p)
		c.deps(c.GopathSrc(p), p)
		hadScaned.Store(subDir, true)
	}

	// 获取_test.go中的依赖包信息
	testImports := pkg.TestImports
	testImports = append(testImports, pkg.XTestImports...)
	for _, p := range testImports {
		if _, ok := hadScaned.Load(p); ok {
			continue
		}
		if p == "C" || build.IsLocalImport(p) {
			continue
		}
		if ok, _ := c.buildInPackage(p); ok {
			continue
		}
		if strings.HasPrefix(p, module) {
			goto CON2
		}
		// 基础库所依赖的三方包
		if baseLib(module, baseModulePath) && !baseLib(p, baseModulePath) {
			lock.Lock()
			base1deps[p] = true
			lock.Unlock()
		} else if baseLib(module, baseModulePath2) && !baseLib(p, baseModulePath2) {
			lock.Lock()
			base2deps[p] = true
			lock.Unlock()
		} else {
			lock.Lock()
			others[p] = true
			lock.Unlock()
		}
	CON2:
		subDir := c.GopathSrc(p)
		hadScaned.Store(subDir, true)
	}

	return nil
}

func (c *Context) buildInPackage(path string) (bool, error) {
	full := filepath.Join(c.GOROOT, "src", path)
	exist, err := IsPathExist(full)
	if err != nil {
		return false, err
	}
	return exist, nil
}

func (c *Context) GopathSrc(path string) string {
	return filepath.Join(c.GOPATH, "src", path)
}

func (c *Context) setPaths() error {
	var err error
	if c.WorkDir, err = os.Getwd(); err != nil {
		return err
	}
	if len(c.WorkDir) == 0 {
		return errors.New("current WorkDir is empty.")
	}
	goPath := os.Getenv("GOPATH")
	GOPATHs := filepath.SplitList(goPath)
	if len(GOPATHs) > 1 {
		return fmt.Errorf("Oops, GOPATH should only one, but, multi GOPATH, %v", GOPATHs)
	}
	if len(GOPATHs) == 0 {
		return errors.New("GOPATH is an empty path, Please check your GOPATH config.")
	}
	c.GOPATH = goPath
	c.GOROOT = build.Default.GOROOT
	if c.GOROOT == "" {
		return errors.New("GOROOT is an empty path, Please check your GOROOT config.")
	}
	return nil
}

func (c *Context) simplify(project string) {
	for k := range others {
		if strings.Contains(c.WorkDir, k) || strings.HasPrefix(k, project) {
			delete(others, k)
			continue
		}
		if baseLib(k, baseModulePath) {
			base1[k] = true
			delete(others, k)
		} else if baseLib(k, baseModulePath2) {
			base2[k] = true
			delete(others, k)
		}
	}
	slim(base1)
	slim(base2)
	slim(base1deps)
	slim(base2deps)
	slim(others)
}

func slim(data map[string]bool) {
	pkgs := make([]string, 0, len(data))
	for k := range data {
		pkgs = append(pkgs, k)
		delete(data, k)
	}
	sort.Strings(pkgs)
	realPath := map[string]struct{}{}
	for _, p := range pkgs {
		found := false
		for r := range realPath {
			if strings.HasPrefix(p, r) {
				found = true
				break
			}
		}
		if found {
			continue
		}
		pp, _ := exchagneModPath(p)
		if len(pp) > 0 {
			realPath[pp] = struct{}{}
		}
		data[pp] = true
	}
}

func (c *Context) loadDeps(runPath, project string, reScan bool) map[string]*Dependence {
	depsMap := map[string]*Dependence{}
	deps, err := c.ScanDeps(runPath, project, reScan)
	if err != nil {
		return nil
	}
	crt := NewConcurrent()
	for _, dep := range deps {
		if dep == project {
			continue
		}
		dep := dep
		cls := func() (interface{}, error) {
			d, err := saveDepObject(dep, c)
			return d, err
		}
		crt.Do(cls)
	}
	res, _ := crt.DoDone()
	for _, d := range res {
		if dep, ok := d.(*Dependence); ok && dep != nil {
			if _, ok := depsMap[dep.ImportPath]; !ok {
				depsMap[dep.ImportPath] = dep
			}
		}
	}
	return depsMap
}
