package main

import (
	"bufio"
	"os"
	"strings"
)

var cmdRecover = &Command{
	Run:   runRecover,
	Name:  "recover",
	Short: "将之前build时候bak的依赖包恢复,保证之前本地的修改不丢失.",
}

func runRecover(args []string, ctx *Context) error {
	recoverFile := ctx.WorkDir + "/.recover"
	f, e1 := os.OpenFile(recoverFile, os.O_RDONLY, 0666)
	if e1 != nil {
		return e1
	}
	defer func() {
		f.Close()
		Remove(recoverFile)
	}()

	r := bufio.NewReader(f)
	for {
		line, _, err := r.ReadLine()
		if err != nil && strings.Contains(err.Error(), "EOF") {
			break
		}
		if len(line) > 0 {
			ss := strings.Split(string(line), ",")
			if len(ss) == 2 {
				src := ss[0]
				dst := ss[1]
				Remove(src)
				Move(dst, src)
			}
		}
	}
	return nil
}
