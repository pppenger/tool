package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
)

type Entry struct {
	Type   string `json:"type"`
	Secret string `json:"secret"`
}

type Conf struct {
	Program    string           `json:"program"`
	ServerCert string           `json:"server_cert"`
	Server     string           `json:"server"`
	UserName   string           `json:"user_name"`
	FormEntry  map[string]Entry `json:"form_entry"`
}

func ParseConf(file string) (*Conf, error) {
	data, err := ioutil.ReadFile(file)
	if err != nil {
		return nil, fmt.Errorf("ioutil.ReadFile err: %w", err)
	}

	var conf Conf
	if err := json.Unmarshal(data, &conf); err != nil {
		return nil, fmt.Errorf("json.Unmarshal err: %w", err)
	}

	return &conf, nil
}
