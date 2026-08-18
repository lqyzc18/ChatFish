// Package config 负责应用配置的加载与持久化。
//
// 设计目标：
// - 与现有字段兼容（api_key/base_url/model）
// - 读取缺失文件时返回空配置（便于首次启动）
// - 保存使用原子写入（临时文件 + Rename），降低写入过程中崩溃导致的配置损坏风险
package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

var userConfigDir = os.UserConfigDir

type Config struct {
	APIKey  string `yaml:"api_key"`
	BaseURL string `yaml:"base_url,omitempty"`
	Model   string `yaml:"model,omitempty"`
}

// Normalize removes accidental leading and trailing whitespace from text fields.
func (c *Config) Normalize() {
	// 规范化输入：去掉用户手动输入或复制粘贴时带来的前后空白。
	if c == nil {
		return
	}
	c.APIKey = strings.TrimSpace(c.APIKey)
	c.BaseURL = strings.TrimSpace(c.BaseURL)
	c.Model = strings.TrimSpace(c.Model)
}

func Load() (*Config, error) {
	// 若 config.yaml 不存在，返回空配置并让 UI 引导用户去设置 API Key。
	cfgPath, err := getConfigPath()
	if err != nil {
		return nil, err
	}

	if _, err := os.Stat(cfgPath); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return &Config{}, nil
		}
		return nil, fmt.Errorf("stat config file %s: %w", cfgPath, err)
	}

	data, err := os.ReadFile(cfgPath)
	if err != nil {
		return nil, fmt.Errorf("read config file %s: %w", cfgPath, err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parse config: %w", err)
	}
	cfg.Normalize()

	return &cfg, nil
}

func Save(cfg *Config) error {
	if cfg == nil {
		return errors.New("config cannot be nil")
	}

	cfgPath, err := getConfigPath()
	if err != nil {
		return err
	}

	dir := filepath.Dir(cfgPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("create config directory: %w", err)
	}

	normalized := *cfg
	normalized.Normalize()
	data, err := yaml.Marshal(&normalized)
	if err != nil {
		return fmt.Errorf("marshal config: %w", err)
	}

	// 使用临时文件 + Rename 实现“原子写入”：避免写入过程中程序崩溃导致配置损坏。
	tempFile, err := os.CreateTemp(dir, ".config-*.tmp")
	if err != nil {
		return fmt.Errorf("create temporary config file: %w", err)
	}
	tempPath := tempFile.Name()
	defer os.Remove(tempPath)

	if err := tempFile.Chmod(0600); err != nil {
		tempFile.Close()
		return fmt.Errorf("set config file permissions: %w", err)
	}
	if _, err := tempFile.Write(data); err != nil {
		tempFile.Close()
		return fmt.Errorf("write temporary config file: %w", err)
	}
	if err := tempFile.Sync(); err != nil {
		tempFile.Close()
		return fmt.Errorf("sync temporary config file: %w", err)
	}
	if err := tempFile.Close(); err != nil {
		return fmt.Errorf("close temporary config file: %w", err)
	}
	if err := os.Rename(tempPath, cfgPath); err != nil {
		return fmt.Errorf("replace config file: %w", err)
	}

	*cfg = normalized
	return nil
}

func getConfigPath() (string, error) {
	configDir, err := userConfigDir()
	if err != nil {
		return "", fmt.Errorf("get user config dir: %w", err)
	}

	return filepath.Join(configDir, "ChatFish", "config.yaml"), nil
}
