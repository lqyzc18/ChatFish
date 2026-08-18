package config

import (
	"os"
	"path/filepath"
	"testing"
)

func useTempConfigDir(t *testing.T) string {
	t.Helper()

	dir := t.TempDir()
	original := userConfigDir
	userConfigDir = func() (string, error) { return dir, nil }
	t.Cleanup(func() { userConfigDir = original })
	return dir
}

func TestLoadReturnsEmptyConfigWhenMissing(t *testing.T) {
	useTempConfigDir(t)

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if *cfg != (Config{}) {
		t.Fatalf("Load() = %#v, want empty config", cfg)
	}
}

func TestSaveAndLoadNormalizesConfig(t *testing.T) {
	useTempConfigDir(t)
	cfg := &Config{
		APIKey:  "  test-key  ",
		BaseURL: " https://example.com/v1 ",
		Model:   " model-name ",
	}

	if err := Save(cfg); err != nil {
		t.Fatalf("Save() error = %v", err)
	}
	if got, want := cfg.APIKey, "test-key"; got != want {
		t.Fatalf("APIKey after Save() = %q, want %q", got, want)
	}

	loaded, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	want := Config{
		APIKey:  "test-key",
		BaseURL: "https://example.com/v1",
		Model:   "model-name",
	}
	if *loaded != want {
		t.Fatalf("Load() = %#v, want %#v", loaded, want)
	}
}

func TestLoadRejectsInvalidYAML(t *testing.T) {
	dir := useTempConfigDir(t)
	path := filepath.Join(dir, "ChatFish", "config.yaml")
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}
	if err := os.WriteFile(path, []byte("api_key: [invalid"), 0600); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	if _, err := Load(); err == nil {
		t.Fatal("Load() error = nil, want invalid YAML error")
	}
}

func TestSaveRejectsNilConfig(t *testing.T) {
	useTempConfigDir(t)

	if err := Save(nil); err == nil {
		t.Fatal("Save(nil) error = nil, want error")
	}
}
