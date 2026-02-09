package parser

import (
	"os"
	"testing"
)

func TestParseStatsCache(t *testing.T) {
	path := DefaultStatsCachePath()
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Skip("stats-cache.json not found, skipping")
	}

	cache, err := ParseStatsCache(path)
	if err != nil {
		t.Fatalf("ParseStatsCache: %v", err)
	}

	if cache.Version == 0 {
		t.Error("expected non-zero version")
	}
	if cache.TotalSessions == 0 {
		t.Error("expected non-zero total sessions")
	}
	if cache.TotalMessages == 0 {
		t.Error("expected non-zero total messages")
	}
	if len(cache.DailyActivity) == 0 {
		t.Error("expected non-empty daily activity")
	}
	if len(cache.ModelUsage) == 0 {
		t.Error("expected non-empty model usage")
	}

	t.Logf("Stats: %d sessions, %d messages, %d models",
		cache.TotalSessions, cache.TotalMessages, len(cache.ModelUsage))
}
