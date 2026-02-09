package parser

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/isaacaudet/clawdtop/internal/model"
)

// DefaultStatsCachePath returns the default path to stats-cache.json.
func DefaultStatsCachePath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(home, ".claude", "stats-cache.json")
}

// ParseStatsCache reads and parses the stats-cache.json file.
func ParseStatsCache(path string) (*model.StatsCache, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading stats cache: %w", err)
	}

	var cache model.StatsCache
	if err := json.Unmarshal(data, &cache); err != nil {
		return nil, fmt.Errorf("parsing stats cache: %w", err)
	}

	return &cache, nil
}
