package provider

import (
	"os"
	"time"

	"github.com/isaacaudet/clawdtop/internal/model"
	"github.com/isaacaudet/clawdtop/internal/parser"
)

// Claude implements Provider for Claude Code / OpenClaw.
type Claude struct {
	StatsPath   string
	ProjectsDir string
}

func NewClaude() *Claude {
	return &Claude{
		StatsPath:   parser.DefaultStatsCachePath(),
		ProjectsDir: parser.DefaultProjectsDir(),
	}
}

func (c *Claude) Name() string  { return "Claude Code" }
func (c *Claude) Icon() string  { return "◈" }
func (c *Claude) Color() string { return "#b4befe" } // Lavender

func (c *Claude) Available() bool {
	_, err := os.Stat(c.StatsPath)
	return err == nil
}

func (c *Claude) Load() (*ProviderData, error) {
	cache, err := parser.ParseStatsCache(c.StatsPath)
	if err != nil {
		return nil, err
	}

	data := &ProviderData{
		ProviderName: c.Name(),
		Icon:         c.Icon(),
		Color:        c.Color(),
		Metadata:     make(map[string]string),
	}

	// Compute total cost from model usage.
	data.TotalCost = model.TotalCostFromModelUsage(cache.ModelUsage)

	// Build model breakdowns.
	for name, mu := range cache.ModelUsage {
		cost := model.CalculateCostFromModelUsage(name, mu)
		data.Models = append(data.Models, ModelBreakdown{
			Model:        model.NormalizeModelName(name),
			InputTokens:  mu.InputTokens,
			OutputTokens: mu.OutputTokens,
			CacheRead:    mu.CacheReadInputTokens,
			CacheWrite:   mu.CacheCreationInputTokens,
			Cost:         cost,
		})
	}

	// Build daily usage.
	days := model.AggregateDaily(cache)
	for _, d := range days {
		data.DailyUsage = append(data.DailyUsage, DailyUsage{
			Date:     d.Date,
			Cost:     d.Cost,
			Tokens:   d.TotalTokens,
			Messages: d.Messages,
			Sessions: d.Sessions,
		})
	}

	// Parse first/last dates.
	if cache.FirstSessionDate != "" {
		if t, err := time.Parse(time.RFC3339, cache.FirstSessionDate); err == nil {
			data.FirstSeen = t
		}
	}
	if len(cache.DailyActivity) > 0 {
		last := cache.DailyActivity[len(cache.DailyActivity)-1]
		if t, err := time.Parse("2006-01-02", last.Date); err == nil {
			data.LastSeen = t
		}
	}

	// Load session data in the background-capable way.
	sessions, _ := parser.LoadAllSessions(c.ProjectsDir)
	for _, s := range sessions {
		var cost float64
		var totalTokens int
		for m, mu := range s.Models {
			cost += model.CalculateCost(m, mu)
			totalTokens += mu.InputTokens + mu.OutputTokens + mu.CacheRead + mu.CacheWrite
		}
		data.Sessions = append(data.Sessions, SessionInfo{
			ID:           s.ID,
			Project:      s.Project,
			StartTime:    s.StartTime,
			EndTime:      s.EndTime,
			Messages:     s.MessageCount,
			UserMessages: s.UserMessages,
			Tokens:       totalTokens,
			Cost:         cost,
		})
	}

	return data, nil
}
