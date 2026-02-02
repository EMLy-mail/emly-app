package utils

import (
	"log"
	"os"
	"path/filepath"

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
}

// LoadConfig reads the config.ini file at the given path and returns a Config struct
func LoadConfig(path string) (*Config, error) {
	cfg, err := ini.Load(path)
	if err != nil {
		log.Printf("Fail to read file: %v", err)
		return nil, err
	}

	config := new(Config)
	if err := cfg.MapTo(config); err != nil {
		log.Printf("Fail to map config: %v", err)
		return nil, err
	}

	return config, nil
}

func SaveConfig(path string, config *Config) error {
	cfg := ini.Empty()
	if err := cfg.ReflectFrom(config); err != nil {
		log.Printf("Fail to reflect config: %v", err)
		return err
	}
	if err := cfg.SaveTo(path); err != nil {
		log.Printf("Fail to save config file: %v", err)
		return err
	}
	return nil
}

func DefaultConfigPath() string {
	// Prefer config.ini next to the executable (packaged app), fallback to CWD (dev).
	exe, err := os.Executable()
	if err == nil {
		p := filepath.Join(filepath.Dir(exe), "config.ini")
		if _, statErr := os.Stat(p); statErr == nil {
			return p
		}
	}
	return "config.ini"
}
