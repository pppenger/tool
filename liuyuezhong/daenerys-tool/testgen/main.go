package main

import (
	"flag"
	"fmt"
	"os"

	"git.inke.cn/BackendPlatform/daenerys-tool/testgen/pkg"
)

var (
	err    error
	files  []string
	parses []*pkg.Parse
)

func main() {
	flag.StringVar(&pkg.Fn, "f", "", "Generating code by function.")
	flag.StringVar(&pkg.InitMain, "init", "", "Generating main_test.go [rpc|http|default empty]")
	flag.BoolVar(&pkg.OldStyle, "old", false, "Force to generate old style test function")
	flag.Parse()
	if err = pkg.ParseArgs(os.Args[1:], &files, 0); err != nil || len(files) == 0 {
		os.Args = append(os.Args, ".")
		pkg.ParseArgs(os.Args[1:], &files, 0)
	}
	if parses, err = pkg.ParseFile(files...); err != nil {
		panic(err)
	}
	if err = pkg.GenTest(parses); err != nil {
		panic(err)
	}

	fmt.Println(`Generation finish!`)
}
