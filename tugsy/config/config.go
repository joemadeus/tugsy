package config

import (
	"errors"
	"os"

	"github.com/mgutz/logxi/v1"
	"github.com/spf13/viper"
)

const (
	resourcesDir = "/Resources"
	spritesDir   = "/sprites"
	osxAppDir    = "/Applications/Tugsy.app"
	devAppDir    = "."
)

var logger = log.New("config")

type Config struct {
	*viper.Viper
	resourcesDirectory string
}

func NewConfig() (*Config, error) {
	var resourcesDirectory string
	if _, err := os.Stat(osxAppDir + resourcesDir); err == nil {
		resourcesDirectory = osxAppDir + resourcesDir
	} else if _, err := os.Stat(devAppDir + resourcesDir); err == nil {
		resourcesDirectory = devAppDir + resourcesDir
	} else {
		return nil, errors.New("Could not determine the base dir for the app")
	}

	logger.Info("Loading resources", "path", resourcesDirectory)

	// Anything set in the environment takes precedence over files
	viperConfig := viper.New()
	viperConfig.AutomaticEnv()
	viperConfig.AddConfigPath(resourcesDirectory)
	viperConfig.SetConfigName("config")

	err := viperConfig.ReadInConfig()
	if err != nil {
		return nil, err
	}

	return &Config{viperConfig, resourcesDirectory}, nil
}

// Returns a path to the sprite sheets
func (config *Config) GetSpritesheetPath(spritesFile string) string {
	return config.resourcesDirectory + spritesDir + "/" + spritesFile
}

// Returns a path to the resources for a given view
func (config *Config) GetViewPath(viewName string) string {
	return config.resourcesDirectory + "/" + viewName + "/"
}
