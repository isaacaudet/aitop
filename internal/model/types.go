package model

import "time"

// StatsCache represents the top-level stats-cache.json structure.
type StatsCache struct {
	Version          int              `json:"version"`
	LastComputedDate string           `json:"lastComputedDate"`
	DailyActivity    []DailyActivity  `json:"dailyActivity"`
	DailyModelTokens []DailyModelTokens `json:"dailyModelTokens"`
	ModelUsage       map[string]ModelUsage `json:"modelUsage"`
	TotalSessions    int              `json:"totalSessions"`
	TotalMessages    int              `json:"totalMessages"`
	LongestSession   LongestSession   `json:"longestSession"`
	FirstSessionDate string           `json:"firstSessionDate"`
	HourCounts       map[string]int   `json:"hourCounts"`
}

type DailyActivity struct {
	Date         string `json:"date"`
	MessageCount int    `json:"messageCount"`
	SessionCount int    `json:"sessionCount"`
	ToolCallCount int   `json:"toolCallCount"`
}

type DailyModelTokens struct {
	Date          string         `json:"date"`
	TokensByModel map[string]int `json:"tokensByModel"`
}

type ModelUsage struct {
	InputTokens              int `json:"inputTokens"`
	OutputTokens             int `json:"outputTokens"`
	CacheReadInputTokens     int `json:"cacheReadInputTokens"`
	CacheCreationInputTokens int `json:"cacheCreationInputTokens"`
}

type LongestSession struct {
	SessionID    string `json:"sessionId"`
	Duration     int64  `json:"duration"`
	MessageCount int    `json:"messageCount"`
	Timestamp    string `json:"timestamp"`
}

// SessionMessage represents a single message from a JSONL session file.
type SessionMessage struct {
	Type      string         `json:"type"`
	SessionID string         `json:"sessionId"`
	Timestamp string         `json:"timestamp"`
	UUID      string         `json:"uuid"`
	Message   *MessageDetail `json:"message,omitempty"`
	CWD       string         `json:"cwd"`
}

type MessageDetail struct {
	Role    string `json:"role"`
	Model   string `json:"model"`
	Usage   *Usage `json:"usage,omitempty"`
}

type Usage struct {
	InputTokens              int `json:"input_tokens"`
	OutputTokens             int `json:"output_tokens"`
	CacheReadInputTokens     int `json:"cache_read_input_tokens"`
	CacheCreationInputTokens int `json:"cache_creation_input_tokens"`
}

// Session is an aggregated view of a single session.
type Session struct {
	ID           string
	Project      string
	StartTime    time.Time
	EndTime      time.Time
	MessageCount int
	UserMessages int
	TokenUsage   TokenUsage
	Models       map[string]TokenUsage
}

// TokenUsage holds aggregated token counts.
type TokenUsage struct {
	InputTokens  int
	OutputTokens int
	CacheRead    int
	CacheWrite   int
}

// DailyStats holds aggregated stats for a single day.
type DailyStats struct {
	Date          string
	Messages      int
	Sessions      int
	ToolCalls     int
	TotalTokens   int
	TokensByModel map[string]int
	Cost          float64
}

// PeriodSummary holds aggregated stats for a time period.
type PeriodSummary struct {
	Label       string
	Days        int
	Messages    int
	Sessions    int
	ToolCalls   int
	TotalTokens int
	Cost        float64
	ModelCosts  map[string]float64
}

// BurnRate holds spending rate information.
type BurnRate struct {
	DailyAvg        float64
	ProjectedMonth  float64
	TrendVsLastWeek float64 // percentage change
}
