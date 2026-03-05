package utils

import (
	"os"
	"path/filepath"

	"emly/backend/logger"

	"gopkg.in/ini.v1"
)

// Config represents the structure of config.ini
type Config struct {
	EMLy EMLyConfig `ini:"EMLy"`
}

type EMLyConfig struct {
	SDKDecoderSemver         string `ini:"SDK_DECODER_SEMVER"`
	SDKDecoderReleaseChannel string `ini:"SDK_DECODER_RELEASE_CHANNEL"`
	GUISemver                string `ini:"GUI_SEMVER"`
	GUIReleaseChannel        string `ini:"GUI_RELEASE_CHANNEL"`
	Language                 string `ini:"LANGUAGE"`
	UpdateCheckEnabled       string `ini:"UPDATE_CHECK_ENABLED"`
	UpdatePath               string `ini:"UPDATE_PATH"`
	UpdateAutoCheck          string `ini:"UPDATE_AUTO_CHECK"`
	BugReportAPIURL          string `ini:"BUGREPORT_API_URL"`
	BugReportAPIKey          string `ini:"BUGREPORT_API_KEY"`
}

// LoadConfig reads the config.ini file at the given path and returns a Config struct
func LoadConfig(path string) (*Config, error) {
	logger.Log("LoadConfig path:", path)
	cfg, err := ini.Load(path)
	if err != nil {
		logger.Log("Fail to read file:", err)
		return nil, err
	}

	config := new(Config)
	if err := cfg.MapTo(config); err != nil {
		logger.Log("Fail to map config:", err)
		return nil, err
	}

	return config, nil
}

func SaveConfig(path string, config *Config) error {
	logger.Log("SaveConfig path:", path)
	cfg := ini.Empty()
	if err := cfg.ReflectFrom(config); err != nil {
		logger.Log("Fail to reflect config:", err)
		return err
	}
	if err := cfg.SaveTo(path); err != nil {
		logger.Log("Fail to save config file:", err)
		return err
	}
	return nil
}

func DefaultConfigPath() string {
	configName := "config.ini"
	if isDebugBuild {
		logger.Log("Debug build: using config.debug.ini")
		configName = "config.debug.ini"
	}

	// Prefer the config file next to the executable (packaged app), fallback to CWD (dev).
	exe, err := os.Executable()
	if err == nil {
		p := filepath.Join(filepath.Dir(exe), configName)
		if _, statErr := os.Stat(p); statErr == nil {
			return p
		}
	}
	return configName
}
