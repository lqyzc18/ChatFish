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
	cfgPath, err := getConfigPath()
	if err != nil {
		return nil, err
	}

	if _, err := os.Stat(cfgPath); os.IsNotExist(err) {
		return &Config{}, nil
	}

	data, err := os.ReadFile(cfgPath)
	if err != nil {
		return nil, fmt.Errorf("read config file %s: %w", cfgPath, err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parse config: %w", err)
	}

	return &cfg, nil
}

func Save(cfg *Config) error {
	cfgPath, err := getConfigPath()
	if err != nil {
		return err
	}

	dir := filepath.Dir(cfgPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("create config directory: %w", err)
	}

	data, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("marshal config: %w", err)
	}

	if err := os.WriteFile(cfgPath, data, 0644); err != nil {
		return fmt.Errorf("write config file: %w", err)
	}

	return nil
}

func getConfigPath() (string, error) {
	execPath, err := os.Executable()
	if err != nil {
		return "", fmt.Errorf("get executable path: %w", err)
	}

	cfgPath := filepath.Join(filepath.Dir(execPath), "config", "config.yaml")

	if _, err := os.Stat(cfgPath); os.IsNotExist(err) {
		cwd, err := os.Getwd()
		if err != nil {
			return "", fmt.Errorf("get working directory: %w", err)
		}
		cfgPath = filepath.Join(cwd, "config", "config.yaml")
	}

	return cfgPath, nil
}
