package main

import (
	"github.com/spf13/viper"
)

type Config struct {
	*viper.Viper
}

func LoadConfig(configPath string) (*Config, error) {
	cfg := &Config{viper.New()}

	// First load config from environment variables.
	// First load wins, so anything set in the environment takes precedence over files.
	cfg.AutomaticEnv()

	// Next load config from files in /config. See https://github.com/spf13/viper#reading-config-files
	// for details about how this works. The code below makes it look for a file at /config/config.XXX,
	// where XXX can be one of a few different supported extensions.
	cfg.AddConfigPath(configPath)
	cfg.SetConfigName("config")

	err := cfg.ReadInConfig()
	if err != nil {
		return nil, err
	}

	return cfg, nil
}
