package main

import (
	"os"

	"github.com/spf13/viper"
)

const (
	resourcesDir = "/Resources"
	spritesDir   = "/sprites"
	osxAppDir    = "/Applications/Tugsy.app"
	devAppDir    = "."
)

type Config struct {
	*viper.Viper
}

func LoadConfig() (*Config, error) {
	cfg := &Config{viper.New()}

	// Anything set in the environment takes precedence over files
	cfg.AutomaticEnv()
	cfg.AddConfigPath(getResourcesDir())
	cfg.SetConfigName("config")

	err := cfg.ReadInConfig()
	if err != nil {
		return nil, err
	}

	return cfg, nil
}

// Returns the resources dir for the app, preferring the OSX dir over the dev dir
func getResourcesDir() string {
	if _, err := os.Stat(osxAppDir + resourcesDir); err == nil {
		return osxAppDir + resourcesDir
	} else if _, err := os.Stat(devAppDir + resourcesDir); err == nil {
		return devAppDir + resourcesDir
	} else {
		logger.Fatal("Could not determine the base dir for the app")
		return ""
	}
}

// Returns a path to a sprite sheet
func getSpritePath(pngResource string) string {
	return getResourcesDir() + spritesDir + "/" + pngResource
}
