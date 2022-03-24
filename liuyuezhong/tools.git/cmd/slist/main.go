package main

import (
	"flag"
	"fmt"
	"os"
)

var (
	serviceName string
	verbose     bool
	sshOption   string
	envFile     string
	rawMode     bool
)

func init() {
	flag.StringVar(&serviceName, "s", "", "service discovery name")
	flag.BoolVar(&verbose, "v", false, "verbose mode")
	flag.StringVar(&sshOption, "so", "stricthostkeychecking=no", "ssh options")
	flag.StringVar(&envFile, "e", "slist.env", "set env file path")
	flag.BoolVar(&rawMode, "raw", false, "raw mode")
}

func main() {
	flag.Parse()

	if err := InitConfig(envFile); err != nil {
		if !os.IsNotExist(err) {
			fmt.Printf("[init config]: %v\n", err)
		}
	}

	infos, err := QueryService(serviceName)
	if err != nil {
		panic(fmt.Errorf("QueryService: %w", err))
	}

	if len(infos) == 0 {
		panic(fmt.Errorf("QueryService: not found"))
	}

	args := flag.Args()
	var cmd string
	if len(args) > 0 {
		cmd = args[0]
	}

	BatchRun(infos, cmd)
}
