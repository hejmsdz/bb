package main

import (
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
	"github.com/hejmsdz/bb/prs"
	"github.com/kirsle/configdir"
)

type Config struct {
	UpdateIntervalMinutes int
	Bitbucket             prs.AccountConfig
	LocalRepositoryPaths  map[string]string
}

var configDirPath string = configdir.LocalConfig("bb")
var configFilePath string = filepath.Join(configDirPath, "/config.toml")

func ReadConfig() (Config, bool) {
	configdir.MakePath(configDirPath)

	var config Config
	tomlData, err := os.ReadFile(configFilePath)
	if err != nil {
		return Config{}, false
	}
	_, err = toml.Decode(string(tomlData), &config)
	if err != nil {
		return Config{}, false
	}
	return config, true
}

func CreateSampleConfig() error {
	tomlData := ` # How often should the list of pull requests be updated?
UpdateIntervalMinutes = 5

[Bitbucket]
# Your Bitbucket username.
# If you don't remember it because you log in via an identity provider,
# you can check it here: https://bitbucket.org/account/settings/username/change
Username = ""

# An app password with permissions to read your user account data and pull requests.
# To generate an app password, go to: https://bitbucket.org/account/settings/app-passwords/new
Password = ""

# Which repositories do you want to monitor?
Repositories = [
	# "owner/reponame",
]

# Where are your local copies of these repositories?
# Configuring these paths is optional, but it will allow you
# to quickly checkout and update branches directly from the app.
[LocalRepositoryPaths]
# "owner/reponame" = "/home/you/Code/reponame"
`
	return os.WriteFile(configFilePath, []byte(tomlData), 0600)
}
