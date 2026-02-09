package provider

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/isaacaudet/aitop/internal/model"
)

// Codex implements Provider for OpenAI Codex CLI.
type Codex struct {
	SessionsDir string
	HistoryPath string
}

func NewCodex() *Codex {
	home, _ := os.UserHomeDir()
	return &Codex{
		SessionsDir: filepath.Join(home, ".codex", "sessions"),
		HistoryPath: filepath.Join(home, ".codex", "history.jsonl"),
	}
}

func (c *Codex) Name() string  { return "Codex" }
func (c *Codex) Icon() string  { return "⊡" }
func (c *Codex) Color() string { return "#a6e3a1" } // Green

func (c *Codex) Available() bool {
	_, err := os.Stat(c.SessionsDir)
	return err == nil
}

// codexSessionMeta is the first line of a new-format rollout JSONL.
type codexSessionMeta struct {
	Timestamp string `json:"timestamp"`
	Type      string `json:"type"`
	Payload   struct {
		ID            string `json:"id"`
		Timestamp     string `json:"timestamp"`
		CWD           string `json:"cwd"`
		CLIVersion    string `json:"cli_version"`
		Source        string `json:"source"`
		ModelProvider string `json:"model_provider"`
	} `json:"payload"`
}

// codexOldMeta is the first line of an old-format rollout JSONL.
type codexOldMeta struct {
	ID        string `json:"id"`
	Timestamp string `json:"timestamp"`
}

// codexTurnContext contains the model name in new-format sessions.
type codexTurnContext struct {
	Type    string `json:"type"`
	Payload struct {
		CWD   string `json:"cwd"`
		Model string `json:"model"`
	} `json:"payload"`
}

// codexEventMsg wraps event messages including token_count.
type codexEventMsg struct {
	Timestamp string `json:"timestamp"`
	Type      string `json:"type"`
	Payload   struct {
		Type string          `json:"type"`
		Info json.RawMessage `json:"info"`
	} `json:"payload"`
}

// codexTokenInfo contains cumulative token usage from token_count events.
type codexTokenInfo struct {
	TotalTokenUsage struct {
		InputTokens           int `json:"input_tokens"`
		CachedInputTokens     int `json:"cached_input_tokens"`
		OutputTokens          int `json:"output_tokens"`
		ReasoningOutputTokens int `json:"reasoning_output_tokens"`
		TotalTokens           int `json:"total_tokens"`
	} `json:"total_token_usage"`
}

// codexResponseItem represents a response_item line.
type codexResponseItem struct {
	Type    string `json:"type"`
	Payload struct {
		Type string `json:"type"`
		Role string `json:"role"`
	} `json:"payload"`
}

// codexOldMessage represents a message line in old-format files.
type codexOldMessage struct {
	Type string `json:"type"`
	Role string `json:"role"`
}

// codexHistoryEntry represents a line in history.jsonl.
type codexHistoryEntry struct {
	SessionID string `json:"session_id"`
	Ts        int64  `json:"ts"`
	Text      string `json:"text"`
}

// codexSession holds parsed data for a single rollout file.
type codexSession struct {
	id           string
	project      string
	modelName    string
	startTime    time.Time
	endTime      time.Time
	messages     int
	userMessages int
	tokens       codexTokenInfo
	dateKey      string // YYYY-MM-DD from directory path
}

func (c *Codex) Load() (*ProviderData, error) {
	data := &ProviderData{
		ProviderName: c.Name(),
		Icon:         c.Icon(),
		Color:        c.Color(),
		Metadata:     make(map[string]string),
	}

	sessions, err := c.parseSessions()
	if err != nil {
		return nil, fmt.Errorf("parsing codex sessions: %w", err)
	}

	// Aggregate by model and by day.
	modelMap := make(map[string]*ModelBreakdown)
	dailyMap := make(map[string]*DailyUsage)

	for _, s := range sessions {
		tu := s.tokens.TotalTokenUsage
		inputTokens := tu.InputTokens
		cachedTokens := tu.CachedInputTokens
		outputTokens := tu.OutputTokens + tu.ReasoningOutputTokens
		totalTokens := inputTokens + outputTokens

		cost := model.CalculateCost(s.modelName, model.TokenUsage{
			InputTokens:  inputTokens,
			OutputTokens: outputTokens,
			CacheRead:    cachedTokens,
		})

		data.TotalCost += cost

		// Model breakdown.
		mb, ok := modelMap[s.modelName]
		if !ok {
			mb = &ModelBreakdown{Model: s.modelName}
			modelMap[s.modelName] = mb
		}
		mb.InputTokens += inputTokens
		mb.OutputTokens += outputTokens
		mb.CacheRead += cachedTokens
		mb.Cost += cost

		// Daily usage.
		day, ok := dailyMap[s.dateKey]
		if !ok {
			day = &DailyUsage{Date: s.dateKey}
			dailyMap[s.dateKey] = day
		}
		day.Cost += cost
		day.Tokens += totalTokens
		day.Messages += s.messages
		day.Sessions++

		// Session info.
		data.Sessions = append(data.Sessions, SessionInfo{
			ID:           s.id,
			Project:      s.project,
			StartTime:    s.startTime,
			EndTime:      s.endTime,
			Messages:     s.messages,
			UserMessages: s.userMessages,
			Tokens:       totalTokens,
			Cost:         cost,
			Model:        s.modelName,
		})

		// Track first/last seen.
		if !s.startTime.IsZero() {
			if data.FirstSeen.IsZero() || s.startTime.Before(data.FirstSeen) {
				data.FirstSeen = s.startTime
			}
			if s.endTime.After(data.LastSeen) {
				data.LastSeen = s.endTime
			} else if s.startTime.After(data.LastSeen) {
				data.LastSeen = s.startTime
			}
		}
	}

	// Convert maps to slices.
	for _, mb := range modelMap {
		data.Models = append(data.Models, *mb)
	}
	for _, du := range dailyMap {
		data.DailyUsage = append(data.DailyUsage, *du)
	}

	// Sort daily usage by date.
	sort.Slice(data.DailyUsage, func(i, j int) bool {
		return data.DailyUsage[i].Date < data.DailyUsage[j].Date
	})

	return data, nil
}

// parseSessions walks the sessions directory and parses each rollout JSONL file.
func (c *Codex) parseSessions() ([]codexSession, error) {
	var sessions []codexSession

	err := filepath.Walk(c.SessionsDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // skip unreadable entries
		}
		if info.IsDir() || !strings.HasSuffix(path, ".jsonl") {
			return nil
		}

		s, err := c.parseRolloutFile(path)
		if err != nil {
			return nil // skip unparseable files
		}
		sessions = append(sessions, s)
		return nil
	})
	if err != nil {
		return nil, err
	}

	return sessions, nil
}

// parseRolloutFile parses a single rollout JSONL file, handling both old and new formats.
func (c *Codex) parseRolloutFile(path string) (codexSession, error) {
	f, err := os.Open(path)
	if err != nil {
		return codexSession{}, err
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024) // 1MB max line

	// Extract date from directory path: .../YYYY/MM/DD/rollout-...
	dateKey := dateKeyFromPath(path)

	var s codexSession
	s.dateKey = dateKey

	var firstTimestamp, lastTimestamp string

	for scanner.Scan() {
		line := scanner.Bytes()
		if len(line) == 0 {
			continue
		}

		// Peek at the type field to decide how to parse.
		var peek struct {
			Type    string `json:"type"`
			ID      string `json:"id"`
			Role    string `json:"role"`
		}
		if err := json.Unmarshal(line, &peek); err != nil {
			continue
		}

		switch {
		case peek.Type == "session_meta":
			// New format session metadata.
			var meta codexSessionMeta
			if err := json.Unmarshal(line, &meta); err == nil {
				s.id = meta.Payload.ID
				s.project = meta.Payload.CWD
				firstTimestamp = meta.Payload.Timestamp
				lastTimestamp = firstTimestamp
			}

		case peek.ID != "" && peek.Type == "":
			// Old format: first line has id and timestamp at top level.
			var oldMeta codexOldMeta
			if err := json.Unmarshal(line, &oldMeta); err == nil {
				s.id = oldMeta.ID
				firstTimestamp = oldMeta.Timestamp
				lastTimestamp = firstTimestamp
			}

		case peek.Type == "turn_context":
			// Contains model name.
			var tc codexTurnContext
			if err := json.Unmarshal(line, &tc); err == nil {
				if tc.Payload.Model != "" {
					s.modelName = tc.Payload.Model
				}
				if tc.Payload.CWD != "" {
					s.project = tc.Payload.CWD
				}
			}

		case peek.Type == "event_msg":
			// May contain token_count events.
			var evt codexEventMsg
			if err := json.Unmarshal(line, &evt); err == nil {
				if evt.Timestamp != "" {
					lastTimestamp = evt.Timestamp
				}
				if evt.Payload.Type == "token_count" && evt.Payload.Info != nil {
					var info codexTokenInfo
					if err := json.Unmarshal(evt.Payload.Info, &info); err == nil {
						if info.TotalTokenUsage.TotalTokens > 0 {
							s.tokens = info // cumulative; last one wins
						}
					}
				}
				if evt.Payload.Type == "user_message" {
					s.userMessages++
					s.messages++
				}
				if evt.Payload.Type == "agent_message" {
					s.messages++
				}
			}

		case peek.Type == "response_item":
			// Count user and assistant messages.
			var ri codexResponseItem
			if err := json.Unmarshal(line, &ri); err == nil {
				if ri.Payload.Role == "user" {
					s.userMessages++
					s.messages++
				} else if ri.Payload.Role == "assistant" {
					s.messages++
				}
			}

		case peek.Type == "message" && peek.Role != "":
			// Old format message.
			var msg codexOldMessage
			if err := json.Unmarshal(line, &msg); err == nil {
				if msg.Role == "user" {
					s.userMessages++
					s.messages++
				} else if msg.Role == "assistant" {
					s.messages++
				}
			}
		}
	}

	// Parse timestamps.
	if firstTimestamp != "" {
		if t, err := time.Parse(time.RFC3339Nano, firstTimestamp); err == nil {
			s.startTime = t
		} else if t, err := time.Parse(time.RFC3339, firstTimestamp); err == nil {
			s.startTime = t
		}
	}
	if lastTimestamp != "" {
		if t, err := time.Parse(time.RFC3339Nano, lastTimestamp); err == nil {
			s.endTime = t
		} else if t, err := time.Parse(time.RFC3339, lastTimestamp); err == nil {
			s.endTime = t
		}
	}

	// Default model if not found.
	if s.modelName == "" {
		s.modelName = "codex-unknown"
	}

	return s, nil
}

// dateKeyFromPath extracts YYYY-MM-DD from a path like .../sessions/YYYY/MM/DD/rollout-...
func dateKeyFromPath(path string) string {
	dir := filepath.Dir(path)
	dd := filepath.Base(dir)
	mmDir := filepath.Dir(dir)
	mm := filepath.Base(mmDir)
	yyyyDir := filepath.Dir(mmDir)
	yyyy := filepath.Base(yyyyDir)

	if len(yyyy) == 4 && len(mm) == 2 && len(dd) == 2 {
		return yyyy + "-" + mm + "-" + dd
	}
	return "unknown"
}
