package config

import (
	"encoding/json"
	"os"
	"path/filepath"
)

type TaiConfig struct {
	Model string `json:"model"`
}

const defaultModel = "auto"

func DefaultConfig() TaiConfig {
	return TaiConfig{Model: defaultModel}
}

func ConfigPath() string {
	base := os.Getenv("XDG_CONFIG_HOME")
	if base == "" {
		if d, err := os.UserConfigDir(); err == nil {
			base = d
		} else {
			home, _ := os.UserHomeDir()
			base = filepath.Join(home, ".config")
		}
	}
	return filepath.Join(base, "tai", "config.json")
}

func LoadConfig() TaiConfig {
	data, err := os.ReadFile(ConfigPath())
	if err != nil {
		return DefaultConfig()
	}
	cfg := DefaultConfig()
	if err := json.Unmarshal(data, &cfg); err != nil {
		return DefaultConfig()
	}
	if cfg.Model == "" {
		cfg.Model = defaultModel
	}
	return cfg
}

func SaveConfig(cfg TaiConfig) error {
	path := ConfigPath()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, append(data, '\n'), 0o644)
}
