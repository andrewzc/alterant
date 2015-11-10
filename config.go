package main

import (
	"io/ioutil"

	"gopkg.in/yaml.v2"
)

type config struct {
	actions   map[string]*OrderedMap
	encrypted []string
}

func newConfig() *config {
	return &config{}
}

func loadConfig(file string) (*config, error) {
	bytes, err := ioutil.ReadFile(file)
	if err != nil {
		return nil, err
	}

	cfg := newConfig()

	err = yaml.Unmarshal(bytes, &cfg)
	if err != nil {
		return nil, err
	}

	return cfg, nil
}
