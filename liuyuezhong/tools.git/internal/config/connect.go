package config

import (
	"fmt"
	"github.com/BurntSushi/toml"
)

type Database struct {
	Name   string   `toml:"name"`
	Master string   `toml:"master"`
	Slaves []string `toml:"slaves"`
}

type Config struct {
	Database []Database `toml:"database"`
}

func LoadConfig(file string) (*Config, error) {
	var c Config
	if _, err := toml.DecodeFile(file, &c); err != nil {
		return nil, fmt.Errorf("DecodeFile err: %v", err)
	}

	return &c, nil
}
