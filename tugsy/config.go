package main

import (
	"errors"

	"github.com/spf13/viper"
)

const (
	resourcesDir = "/Resources"
	spritesDir   = resourcesDir + "/sprites"
	osxAppDir    = "/Applications/Tugsy.app"
	devAppDir    = "./"
)

type Config struct {
	*viper.Viper
}

func LoadConfig() (*Config, error) {
	cfg := &Config{viper.New()}

	// First load config from environment variables.
	// First load wins, so anything set in the environment takes precedence over files.
	cfg.AutomaticEnv()

	// Next load config from files in /config. See https://github.com/spf13/viper#reading-config-files
	// for details about how this works. The code below makes it look for a file at /config/config.XXX,
	// where XXX can be one of a few different supported extensions.
	appConfig := []string{osxAppDir + resourcesDir, devAppDir + resourcesDir}
	var resourceDir string
	for _, configDir := range appConfig {
		if exists(configDir) {
			resourceDir = configDir
		}
	}

	if resourceDir == "" {
		return nil, errors.New("Could not locate a resources dir")
	}

	cfg.AddConfigPath(resourceDir)
	cfg.SetConfigName("config")

	err := cfg.ReadInConfig()
	if err != nil {
		return nil, err
	}

	return cfg, nil
}
