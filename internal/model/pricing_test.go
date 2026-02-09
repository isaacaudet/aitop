package model

import (
	"math"
	"testing"
)

func TestNormalizeModelName(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"claude-opus-4-5-20251101", "opus-4-5"},
		{"claude-sonnet-4-5-20250929", "sonnet-4-5"},
		{"claude-haiku-4-5-20251001", "haiku-4-5"},
		{"claude-opus-4-6", "opus-4-6"},
		{"claude-opus-4-5-thinking", "opus-4-5-thinking"},
		{"claude-opus-4-1-20250805", "opus-4-1"},
		{"gpt-4o", "gpt-4o"},
		{"o3", "o3"},
		{"o4-mini", "o4-mini"},
		{"models/gemini-2.5-pro", "gemini-2.5-pro"},
		{"gemini-2.5-flash", "gemini-2.5-flash"},
	}
	for _, tt := range tests {
		got := NormalizeModelName(tt.input)
		if got != tt.expected {
			t.Errorf("NormalizeModelName(%q) = %q, want %q", tt.input, got, tt.expected)
		}
	}
}

func TestGetPricing(t *testing.T) {
	tests := []struct {
		model    string
		wantOK   bool
		wantOut  float64
	}{
		{"claude-opus-4-6", true, 25.0},
		{"claude-opus-4-5-20251101", true, 25.0},
		{"claude-sonnet-4-5-20250929", true, 15.0},
		{"claude-haiku-4-5-20251001", true, 4.0},
		{"claude-opus-4-5-thinking", true, 25.0},
		{"claude-opus-4-1-20250805", true, 75.0},
		{"gpt-4o", true, 10.0},
		{"o3", true, 40.0},
		{"o4-mini", true, 4.40},
		{"models/gemini-2.5-pro", true, 10.0},
		{"gemini-2.5-flash", true, 2.50},
		{"unknown-model", false, 0},
	}
	for _, tt := range tests {
		p, ok := GetPricing(tt.model)
		if ok != tt.wantOK {
			t.Errorf("GetPricing(%q) ok = %v, want %v", tt.model, ok, tt.wantOK)
			continue
		}
		if ok && p.OutputPerMTok != tt.wantOut {
			t.Errorf("GetPricing(%q).OutputPerMTok = %v, want %v", tt.model, p.OutputPerMTok, tt.wantOut)
		}
	}
}

func TestCalculateCost(t *testing.T) {
	usage := TokenUsage{
		InputTokens:  1_000_000,
		OutputTokens: 1_000_000,
		CacheRead:    1_000_000,
		CacheWrite:   1_000_000,
	}
	// opus-4-6: $5 + $25 + $0.50 + $6.25 = $36.75
	cost := CalculateCost("claude-opus-4-6", usage)
	if math.Abs(cost-36.75) > 0.01 {
		t.Errorf("CalculateCost opus-4-6 = %v, want 36.75", cost)
	}

	// sonnet-4-5: $3 + $15 + $0.30 + $3.75 = $22.05
	cost = CalculateCost("claude-sonnet-4-5-20250929", usage)
	if math.Abs(cost-22.05) > 0.01 {
		t.Errorf("CalculateCost sonnet-4-5 = %v, want 22.05", cost)
	}

	// gpt-4o (no cache pricing): $2.50 + $10 = $12.50
	cost = CalculateCost("gpt-4o", usage)
	if math.Abs(cost-12.50) > 0.01 {
		t.Errorf("CalculateCost gpt-4o = %v, want 12.50", cost)
	}
}

func TestCalculateCostFromRealData(t *testing.T) {
	// Real data from the user's stats-cache.json
	usage := map[string]ModelUsage{
		"claude-opus-4-5-20251101": {
			InputTokens:              2800456,
			OutputTokens:             2066904,
			CacheReadInputTokens:     3715502867,
			CacheCreationInputTokens: 235557308,
		},
		"claude-opus-4-6": {
			InputTokens:              112444,
			OutputTokens:             958448,
			CacheReadInputTokens:     623295535,
			CacheCreationInputTokens: 44469146,
		},
		"claude-sonnet-4-5-20250929": {
			InputTokens:              53939,
			OutputTokens:             32237,
			CacheReadInputTokens:     272900592,
			CacheCreationInputTokens: 14790323,
		},
	}
	total := TotalCostFromModelUsage(usage)
	// With corrected Opus pricing ($5/$25 instead of $15/$75), total should be lower
	// Primarily cache tokens: ~3.7B cache read at $0.50/M + ~235M cache write at $6.25/M for opus-4-5
	// ~$1857 cache read + ~$1472 cache write for opus-4-5 alone ≈ $3329
	// Plus opus-4-6 and sonnet contributions
	if total < 50 {
		t.Errorf("TotalCostFromModelUsage = %v, expected > 50", total)
	}
	if total > 50000 {
		t.Errorf("TotalCostFromModelUsage = %v, expected < 50000 (sanity check)", total)
	}
	t.Logf("Total cost from real data: $%.2f", total)
}
