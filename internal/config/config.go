// Package config manages gem-image configuration.
package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
)

// Config holds all gem-image configuration.
type Config struct {
	GCP   GCPConfig   `toml:"gcp"`
	Model ModelConfig `toml:"model"`
}

// GCPConfig holds Google Cloud settings.
type GCPConfig struct {
	Project  string `toml:"project"`
	Location string `toml:"location"`
}

// ModelConfig holds model settings.
type ModelConfig struct {
	Name string `toml:"name"`
}

// Load reads config from the given path, with env var overrides.
// If path is empty, tries the default location (~/.config/gem-image/config.toml).
func Load(path string) (*Config, error) {
	cfg := &Config{
		GCP: GCPConfig{
			Location: "us-central1",
		},
		Model: ModelConfig{
			Name: "gemini-2.5-flash-image",
		},
	}

	if path == "" {
		home, err := os.UserHomeDir()
		if err == nil {
			path = filepath.Join(home, ".config", "gem-image", "config.toml")
		}
	}

	if path != "" {
		if _, err := os.Stat(path); err == nil {
			if _, err := toml.DecodeFile(path, cfg); err != nil {
				return nil, fmt.Errorf("parse config %s: %w", path, err)
			}
		}
	}

	// Env overrides (tool-specific > generic)
	if v := os.Getenv("GEMIMAGE_PROJECT"); v != "" {
		cfg.GCP.Project = v
	} else if v := os.Getenv("GOOGLE_CLOUD_PROJECT"); v != "" {
		cfg.GCP.Project = v
	}
	if v := os.Getenv("GEMIMAGE_LOCATION"); v != "" {
		cfg.GCP.Location = v
	} else if v := os.Getenv("GOOGLE_CLOUD_LOCATION"); v != "" {
		cfg.GCP.Location = v
	}
	if v := os.Getenv("GEMIMAGE_MODEL"); v != "" {
		cfg.Model.Name = v
	}

	if cfg.GCP.Project == "" {
		return nil, fmt.Errorf("GCP project is required: set gcp.project in config or GOOGLE_CLOUD_PROJECT env var")
	}

	return cfg, nil
}

// ApplyFlags overrides config values with CLI flag values.
func (c *Config) ApplyFlags(model string) {
	if model != "" {
		c.Model.Name = model
	}
}
