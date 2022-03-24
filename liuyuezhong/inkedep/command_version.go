package main

import "fmt"

var cmdVersion = &Command{
	Run:   runVersion,
	Name:  "ver",
	Short: "工具版本号,目前版本为version 2.0",
}

func runVersion(args []string, ctx *Context) error {
	fmt.Println("inkedep version 2.0")
	return nil
}
