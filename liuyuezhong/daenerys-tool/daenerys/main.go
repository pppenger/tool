package main

import (
	"git.inke.cn/BackendPlatform/daenerys-tool/daenerys/cmd"
)

func main() {
	go func() {
		_ = cmd.ToolStat()
	}()

	cmd.Execute()
}
