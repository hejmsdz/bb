package main

import (
	"os"

	"github.com/BurntSushi/toml"
)

type AccountConfig struct {
	Username     string
	Password     string
	UserId       string
	Repositories []string
}

type Config struct {
	UpdateIntervalMinutes int
	Bitbucket             AccountConfig
}

func ReadConfig() Config {
	var config Config
	tomlData, _ := os.ReadFile("./config.toml")
	toml.Decode(string(tomlData), &config)
	return config
}
