package main

import (
	"flag"
	"github.com/kardianos/service"
	"log"
)

var (
	verbose bool
	file    = "vpn.json"
	action  string

	svcConfig = &service.Config{
		Name:        "myConnect",
		DisplayName: "my connect",
		Description: "nothing",
	}
)

func init() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	flag.BoolVar(&verbose, "v", false, "verbose mode")
	flag.StringVar(&action, "s", "", "Control the system service")
}

func main() {
	flag.Parse()

	prg := NewProgram(file)
	s, err := service.New(prg, svcConfig)
	if err != nil {
		log.Fatal(err)
	}

	if action != "" {
		if err := service.Control(s, action); err != nil {
			log.Fatal(err)
		}

		return
	}

	if err = s.Run(); err != nil {
		log.Println(err)
	}
}
