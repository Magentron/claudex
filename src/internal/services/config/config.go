// Package config provides configuration file loading and parsing for Claudex.
// It supports loading .claudex.toml files with options for documentation paths
// and file overwrite behavior.
package config

import (
	"os"
	"runtime"
	"strconv"

	"github.com/BurntSushi/toml"
	"github.com/spf13/afero"
)

// ProcessProtection configures runaway process protection and process resource limits
type ProcessProtection struct {
	MaxProcesses       int `toml:"max_processes"`
	RateLimitPerSecond int `toml:"rate_limit_per_second"`
	TimeoutSeconds     int `toml:"timeout_seconds"`
}

// Features controls optional token-consuming features
type Features struct {
	AutodocSessionProgress bool              `toml:"autodoc_session_progress"`
	AutodocSessionEnd      bool              `toml:"autodoc_session_end"`
	AutodocFrequency       int               `toml:"autodoc_frequency"`
	ProcessProtection      ProcessProtection `toml:"process_protection"`
}

type Config struct {
	Doc         []string `toml:"doc"`
	NoOverwrite bool     `toml:"no_overwrite"`
	Features    Features `toml:"features"`
}

// Load loads configuration from the specified path using the provided filesystem
func Load(fs afero.Fs, path string) (*Config, error) {
	config := &Config{
		Doc:         []string{},
		NoOverwrite: false,
		Features: Features{
			AutodocSessionProgress: true,
			AutodocSessionEnd:      true,
			AutodocFrequency:       5,
			ProcessProtection: ProcessProtection{
				MaxProcesses:       runtime.NumCPU() * 2,
				RateLimitPerSecond: 5,
				TimeoutSeconds:     300,
			},
		},
	}

	if _, err := fs.Stat(path); err == nil {
		data, err := afero.ReadFile(fs, path)
		if err != nil {
			return nil, err
		}
		if _, err := toml.Decode(string(data), config); err != nil {
			return nil, err
		}
	}

	applyEnvironmentOverrides(config)
	return config, nil
}

// applyEnvironmentOverrides applies environment variable overrides to the configuration
func applyEnvironmentOverrides(config *Config) {
	if val := os.Getenv("CLAUDEX_MAX_PROCESSES"); val != "" {
		if intVal, err := strconv.Atoi(val); err == nil {
			config.Features.ProcessProtection.MaxProcesses = intVal
		}
	}
	if val := os.Getenv("CLAUDEX_RATE_LIMIT"); val != "" {
		if intVal, err := strconv.Atoi(val); err == nil {
			config.Features.ProcessProtection.RateLimitPerSecond = intVal
		}
	}
	if val := os.Getenv("CLAUDEX_TIMEOUT"); val != "" {
		if intVal, err := strconv.Atoi(val); err == nil {
			config.Features.ProcessProtection.TimeoutSeconds = intVal
		}
	}
}
