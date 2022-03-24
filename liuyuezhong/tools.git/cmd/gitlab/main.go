package main

import (
	"flag"
	"fmt"
	"github.com/xanzy/go-gitlab"
	"log"
	"net/url"
	"os"
	"path/filepath"
)

var (
	configFile string
	clone      bool
	verbose    bool
)

func init() {
	base, _ := os.Executable()
	defaultPath := filepath.Join(filepath.Dir(base), "gitlab.yaml")

	flag.StringVar(&configFile, "c", defaultPath, "config file path")
	flag.BoolVar(&clone, "clone", false, "auto clone project")
	flag.BoolVar(&verbose, "v", false, "verbose mode")
}

func main() {
	flag.Parse()

	if verbose {
		fmt.Printf("load config file from %s\n", configFile)
	}
	conf, err := LoadConfig(configFile)
	if err != nil {
		log.Fatalf("LoadConfig err: %v", err)
	}

	git, err := gitlab.NewClient(conf.Gitlab.Token, gitlab.WithBaseURL(conf.Gitlab.API))
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}

	uri, err := url.Parse(conf.Gitlab.API)
	if err != nil {
		log.Fatalf("inlvaid gitlab api url: %v", err)
	}
	hostName := uri.Hostname()

	for _, g := range conf.Groups {
		CompareGroup(git, g, hostName)
		fmt.Println("")
	}
}
