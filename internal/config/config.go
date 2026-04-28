package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type Config struct {
	APIKey string `yaml:"api_key"`
}

func Load() (*Config, error) {
	execPath, err := os.Executable()
	if err != nil {
		return nil, fmt.Errorf("get executable path: %w", err)
	}
	cfgPath := filepath.Join(filepath.Dir(execPath), "config", "config.yaml")

	if _, err := os.Stat(cfgPath); os.IsNotExist(err) {
		execPath, _ = os.Getwd()
		cfgPath = filepath.Join(execPath, "config", "config.yaml")
	}

	data, err := os.ReadFile(cfgPath)
	if err != nil {
		return nil, fmt.Errorf("read config file %s: %w", cfgPath, err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parse config: %w", err)
	}

	if cfg.APIKey == "" {
		return nil, fmt.Errorf("api_key is empty in config")
	}

	return &cfg, nil
}
