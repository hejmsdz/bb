package main

import (
	"os"

	"github.com/BurntSushi/toml"
	"github.com/hejmsdz/bb/prs"
)

type Config struct {
	UpdateIntervalMinutes int
	Bitbucket             prs.AccountConfig
	LocalRepositoryPaths  map[string]string
}

func ReadConfig() Config {
	var config Config
	tomlData, _ := os.ReadFile("./config.toml")
	toml.Decode(string(tomlData), &config)
	return config
}
