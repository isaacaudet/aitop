package provider

import "time"

// Provider is the interface all AI tool trackers implement.
type Provider interface {
	Name() string
	Icon() string // Unicode icon for TUI display
	Color() string // Hex color for charts
	Load() (*ProviderData, error)
	Available() bool // Whether the data source exists on this machine
}

// ProviderData holds normalized usage data from any AI tool.
type ProviderData struct {
	ProviderName string
	Icon         string
	Color        string
	TotalCost    float64
	DailyUsage   []DailyUsage
	Models       []ModelBreakdown
	Sessions     []SessionInfo
	Generations  int // Code generations (for tools like Cursor)
	FirstSeen    time.Time
	LastSeen     time.Time
	Metadata     map[string]string // Provider-specific info
}

// DailyUsage holds aggregated daily data across providers.
type DailyUsage struct {
	Date        string // YYYY-MM-DD
	Cost        float64
	Tokens      int
	Messages    int
	Sessions    int
	Generations int // For code generation tools
}

// ModelBreakdown holds per-model stats.
type ModelBreakdown struct {
	Model        string
	InputTokens  int
	OutputTokens int
	CacheRead    int
	CacheWrite   int
	Cost         float64
	Generations  int
}

// SessionInfo holds a single session's data.
type SessionInfo struct {
	ID           string
	Project      string
	StartTime    time.Time
	EndTime      time.Time
	Messages     int
	UserMessages int
	Tokens       int
	Cost         float64
	Model        string
}

// AggregatedData holds combined data from all providers.
type AggregatedData struct {
	Providers    []*ProviderData
	TotalCost    float64
	DailyUsage   []DailyUsage // merged across providers
	TotalTokens  int
	TotalSessions int
	FirstSeen    time.Time
	LastSeen     time.Time
}

// LoadAll loads data from all available providers.
func LoadAll(providers []Provider) *AggregatedData {
	agg := &AggregatedData{}
	dailyMap := make(map[string]DailyUsage)

	for _, p := range providers {
		if !p.Available() {
			continue
		}
		data, err := p.Load()
		if err != nil {
			continue
		}

		agg.Providers = append(agg.Providers, data)
		agg.TotalCost += data.TotalCost

		for _, d := range data.DailyUsage {
			existing := dailyMap[d.Date]
			existing.Date = d.Date
			existing.Cost += d.Cost
			existing.Tokens += d.Tokens
			existing.Messages += d.Messages
			existing.Sessions += d.Sessions
			existing.Generations += d.Generations
			dailyMap[d.Date] = existing
		}

		for _, s := range data.Sessions {
			agg.TotalSessions++
			agg.TotalTokens += s.Tokens
		}

		if agg.FirstSeen.IsZero() || (!data.FirstSeen.IsZero() && data.FirstSeen.Before(agg.FirstSeen)) {
			agg.FirstSeen = data.FirstSeen
		}
		if data.LastSeen.After(agg.LastSeen) {
			agg.LastSeen = data.LastSeen
		}
	}

	// Convert dailyMap to sorted slice.
	for _, d := range dailyMap {
		agg.DailyUsage = append(agg.DailyUsage, d)
	}
	sortDailyUsage(agg.DailyUsage)

	return agg
}

func sortDailyUsage(days []DailyUsage) {
	for i := 1; i < len(days); i++ {
		for j := i; j > 0 && days[j].Date < days[j-1].Date; j-- {
			days[j], days[j-1] = days[j-1], days[j]
		}
	}
}
