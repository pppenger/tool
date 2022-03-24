package main

import (
	"fmt"
	"gopkg.in/yaml.v2"
	"os"
	"regexp"
)

type Config struct {
	IgnoreHost []string `yaml:"ignore_host"`
	SrvUrl     string   `yaml:"srv_url"`
}

var config *Config

func InitConfig(file string) error {
	data, err := os.ReadFile(file)
	if err != nil {
		return fmt.Errorf("ioutil.ReadFile err: %v", err)
	}

	var conf Config
	if err := yaml.Unmarshal(data, &conf); err != nil {
		return fmt.Errorf("yaml.Unmarshal err: %v", err)
	}
	config = &conf

	return nil
}

func IgnoreHost(host string) bool {
	if config == nil {
		return false
	}

	for _, r := range config.IgnoreHost {
		matched, err := regexp.MatchString(r, host)
		if err != nil {
			fmt.Printf("[IgnoreHost]re(%s) err: %v\n", r, err)
			continue
		}

		if matched {
			return true
		}
	}

	return false
}
