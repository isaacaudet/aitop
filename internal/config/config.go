package config

import (
	"os"
	"path/filepath"

	"github.com/pelletier/go-toml/v2"
)

// PlanConfig holds subscription plan details.
type PlanConfig struct {
	Provider    string  `toml:"provider"`
	Name        string  `toml:"name"`
	MonthlyCost float64 `toml:"monthly_cost"`
}

// Config holds application configuration.
type Config struct {
	StatsCachePath string     `toml:"stats_cache_path"`
	ProjectsDir    string     `toml:"projects_dir"`
	Plan           PlanConfig `toml:"plan"`
}

// DefaultConfigPath returns the path to the config file.
func DefaultConfigPath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(home, ".config", "clawdtop", "config.toml")
}

// Load reads the config file, returning defaults if it doesn't exist.
func Load() Config {
	cfg := Config{
		Plan: PlanConfig{
			Provider:    "claude",
			Name:        "Max",
			MonthlyCost: 200,
		},
	}

	path := DefaultConfigPath()
	data, err := os.ReadFile(path)
	if err != nil {
		return cfg
	}

	_ = toml.Unmarshal(data, &cfg)
	return cfg
}
