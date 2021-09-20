package main

import (
	"encoding/json"
	"os"
)

type (
	config struct {
		Forward []forwardConfig `json:"forward"`
	}
	forwardConfig struct {
		Target string `json:"target"`
		Port   int    `json:"port"`
	}
)

func readConfigFile(filepath string) (config, error) {
	var conf config

	file, err := os.Open(filepath)
	if err != nil {
		return conf, err
	}
	return conf, json.NewDecoder(file).Decode(&conf)
}
