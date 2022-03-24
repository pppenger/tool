package main

import "strings"

var cmdGet = &Command{
	Run:   runGet,
	Name:  "get",
	Short: "下载依赖,类似于go get,优先从公司的gitcache上下载,如果没有则从官网下载.",
}

func runGet(args []string, ctx *Context) error {
	for _, p := range args {
		path := p
		if r, ok := exchagneModPath(path); ok {
			path = r
		} else {
			doPrint("import path(%s) maybe incorrect !!!", path)
			return nil
		}
		dir := ctx.GopathSrc(path)
		if exist, err := IsPathExist(dir); err != nil {
			doPrint("creating repository %s error: %v\n", path, err)
			continue
		} else if exist {
			err := Remove(dir)
			if err != nil {
				return err
			}
		}
		doPrint("use master, creating %s at $GOPATH/src\n", path)
		dep := &Dependence{
			ImportPath: path,
			Rev:        "master",
		}

		// 灰度设置
		setGrayConfig(ctx.WorkDir, dep)

		if strings.HasPrefix(dep.ImportPath, "github.com/apache/thrift") ||
			strings.HasPrefix(dep.ImportPath, "git.apache.org/thrift.git") {
			dep.Rev = "0.10.0"
		}

		d := ctx.GopathSrc(dep.ImportPath)
		err := dep.Create(d)
		if err != nil {
			return err
		}
	}
	return nil
}
