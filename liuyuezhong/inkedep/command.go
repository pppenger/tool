package main

import (
	"bufio"
	"fmt"
	"os"
)

type Command struct {
	Run          func(args []string, ctx *Context) error
	Name         string
	Short        string
	OnlyInGOPATH bool
}

var commands = []*Command{
	cmdGet,
	cmdSave,
	cmdBuild,
	cmdVersion,
	cmdRecover,
}

func Run(cmd string, args []string) error {
	ctx, err := NewCtx()
	if err != nil {
		return err
	}
	for _, c := range commands {
		if c.Name == cmd {
			e := c.Run(args, ctx)
			if cmd == "build" && len(recovering) > 0 {
				recoverFile := ctx.WorkDir + "/.recover"
				f, e1 := os.OpenFile(recoverFile, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
				if e1 != nil {
					return e1
				}
				defer f.Close()

				w := bufio.NewWriter(f)
				for src, dst := range recovering {
					s := fmt.Sprintf("%s,%s\n", src, dst)
					w.Write([]byte(s))
				}
				w.Flush()
			}
			return e
		}
	}

	go func() {
		_ = ToolStat()
	}()

	return fmt.Errorf("inkedep command %s not exist", cmd)
}
