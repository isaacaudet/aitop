package model

import (
	"regexp"
	"strings"
)

// ModelPricing holds per-million-token prices.
type ModelPricing struct {
	InputPerMTok      float64
	OutputPerMTok     float64
	CacheReadPerMTok  float64
	CacheWritePerMTok float64
}

// pricingTable maps model prefix to pricing.
var pricingTable = map[string]ModelPricing{
	// Anthropic Claude models
	"opus-4-6": {
		InputPerMTok:      5.0,
		OutputPerMTok:     25.0,
		CacheReadPerMTok:  0.50,
		CacheWritePerMTok: 6.25,
	},
	"opus-4-5": {
		InputPerMTok:      5.0,
		OutputPerMTok:     25.0,
		CacheReadPerMTok:  0.50,
		CacheWritePerMTok: 6.25,
	},
	"sonnet-4-5": {
		InputPerMTok:      3.0,
		OutputPerMTok:     15.0,
		CacheReadPerMTok:  0.30,
		CacheWritePerMTok: 3.75,
	},
	"haiku-4-5": {
		InputPerMTok:      0.80,
		OutputPerMTok:     4.0,
		CacheReadPerMTok:  0.08,
		CacheWritePerMTok: 1.0,
	},
	"opus-4-1": {
		InputPerMTok:      15.0,
		OutputPerMTok:     75.0,
		CacheReadPerMTok:  1.50,
		CacheWritePerMTok: 18.75,
	},
	// OpenAI models
	"gpt-4o": {
		InputPerMTok:  2.50,
		OutputPerMTok: 10.0,
	},
	"gpt-4.1": {
		InputPerMTok:  2.0,
		OutputPerMTok: 8.0,
	},
	"o3": {
		InputPerMTok:  10.0,
		OutputPerMTok: 40.0,
	},
	"o4-mini": {
		InputPerMTok:  1.10,
		OutputPerMTok: 4.40,
	},
	// Google Gemini models
	"gemini-2.5-pro": {
		InputPerMTok:  1.25,
		OutputPerMTok: 10.0,
	},
	"gemini-2.5-flash": {
		InputPerMTok:  0.30,
		OutputPerMTok: 2.50,
	},
}

// dateSuffixRe strips date suffixes like -20251101 or -20250929.
var dateSuffixRe = regexp.MustCompile(`-\d{8}$`)

// NormalizeModelName strips the "claude-" prefix and date suffix to get the pricing key.
// Also handles OpenAI and Gemini model names like "models/gemini-2.5-pro".
func NormalizeModelName(model string) string {
	name := strings.TrimPrefix(model, "claude-")
	// Strip Gemini API prefix "models/"
	name = strings.TrimPrefix(name, "models/")
	name = dateSuffixRe.ReplaceAllString(name, "")
	return name
}

// GetPricing returns the pricing for a model name (with or without prefix/suffix).
func GetPricing(model string) (ModelPricing, bool) {
	key := NormalizeModelName(model)
	// Try exact match first.
	if p, ok := pricingTable[key]; ok {
		return p, true
	}
	// Try prefix match (e.g., "opus-4-5-thinking" matches "opus-4-5").
	for prefix, p := range pricingTable {
		if strings.HasPrefix(key, prefix) {
			return p, true
		}
	}
	return ModelPricing{}, false
}

// CalculateCost computes the dollar cost for a token usage breakdown.
func CalculateCost(model string, usage TokenUsage) float64 {
	pricing, ok := GetPricing(model)
	if !ok {
		return 0
	}
	cost := float64(usage.InputTokens) * pricing.InputPerMTok / 1_000_000
	cost += float64(usage.OutputTokens) * pricing.OutputPerMTok / 1_000_000
	cost += float64(usage.CacheRead) * pricing.CacheReadPerMTok / 1_000_000
	cost += float64(usage.CacheWrite) * pricing.CacheWritePerMTok / 1_000_000
	return cost
}

// CalculateCostFromModelUsage computes cost from a ModelUsage struct.
func CalculateCostFromModelUsage(model string, mu ModelUsage) float64 {
	return CalculateCost(model, TokenUsage{
		InputTokens:  mu.InputTokens,
		OutputTokens: mu.OutputTokens,
		CacheRead:    mu.CacheReadInputTokens,
		CacheWrite:   mu.CacheCreationInputTokens,
	})
}

// TotalCostFromModelUsage computes total cost across all models.
func TotalCostFromModelUsage(usage map[string]ModelUsage) float64 {
	var total float64
	for model, mu := range usage {
		total += CalculateCostFromModelUsage(model, mu)
	}
	return total
}
