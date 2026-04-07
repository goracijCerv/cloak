package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

type AppConfig struct {
	DefaultOutputDir  string `json:"default_output_dir"`
	AlwaysSkipConfirm bool   `json:"always_skip_confirm"`
	DefaultJSONOutput bool   `json:"default_json_output"`
}

func Load() (*AppConfig, error) {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return nil, fmt.Errorf("could not find config directory: %w", err)
	}

	cloakDir := filepath.Join(configDir, "cloak")
	configPath := filepath.Join(cloakDir, "config.json")
	// #nosec G304 -- This tool needs to read arbitrary files by design
	data, err := os.ReadFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			return createDefaultConfig(cloakDir, configPath)
		}

		return nil, err
	}

	var cfg AppConfig
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}

func createDefaultConfig(cloakDir, configPath string) (*AppConfig, error) {
	if err := os.MkdirAll(cloakDir, 0750); err != nil {
		return nil, err
	}

	defaultConfig := AppConfig{
		DefaultOutputDir:  "",
		AlwaysSkipConfirm: false,
		DefaultJSONOutput: false,
	}

	jsonData, err := json.MarshalIndent(defaultConfig, "", " ")
	if err != nil {
		return nil, err
	}

	if err := os.WriteFile(configPath, jsonData, 0600); err != nil {
		return nil, err
	}

	return &defaultConfig, nil
}
