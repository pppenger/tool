package main

import (
	"gopkg.in/yaml.v2"
	"io/ioutil"
)

type Group struct {
	BaseGroup  string   `yaml:"base_group"`
	IgnoreDirs []string `yaml:"ignore_dirs"`
}

type Config struct {
	Gitlab struct {
		API   string `yaml:"api"`
		Token string `yaml:"token"`
	} `yaml:"gitlab"`

	Groups []Group `yaml:"groups"`
}

func LoadConfig(file string) (*Config, error) {
	data, err := ioutil.ReadFile(file)
	if err != nil {
		return nil, err
	}

	var c Config
	if err := yaml.Unmarshal(data, &c); err != nil {
		return nil, err
	}

	return &c, nil
}
