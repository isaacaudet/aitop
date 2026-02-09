package provider

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/isaacaudet/clawdtop/internal/model"
)

// Gemini implements Provider for Google Gemini CLI.
type Gemini struct {
	ConfigDir string
}

func NewGemini() *Gemini {
	home, _ := os.UserHomeDir()
	return &Gemini{
		ConfigDir: filepath.Join(home, ".gemini"),
	}
}

func (g *Gemini) Name() string  { return "Gemini" }
func (g *Gemini) Icon() string  { return "✦" }
func (g *Gemini) Color() string { return "#74c7ec" } // Sapphire

func (g *Gemini) Available() bool {
	// Check for tmp directory with chat sessions first.
	tmpDir := filepath.Join(g.ConfigDir, "tmp")
	if entries, err := os.ReadDir(tmpDir); err == nil {
		for _, e := range entries {
			if e.IsDir() {
				chatsDir := filepath.Join(tmpDir, e.Name(), "chats")
				if _, err := os.Stat(chatsDir); err == nil {
					return true
				}
			}
		}
	}
	// Fallback: settings.json indicates Gemini is installed.
	_, err := os.Stat(filepath.Join(g.ConfigDir, "settings.json"))
	return err == nil
}

// geminiSession represents the JSON structure of a Gemini session file.
type geminiSession struct {
	SessionID   string          `json:"sessionId"`
	ProjectHash string          `json:"projectHash"`
	StartTime   string          `json:"startTime"`
	LastUpdated string          `json:"lastUpdated"`
	Messages    []geminiMessage `json:"messages"`
}

// geminiMessage represents a single message in a Gemini session.
type geminiMessage struct {
	Type      string        `json:"type"`
	Content   string        `json:"content"`
	Tokens    *geminiTokens `json:"tokens,omitempty"`
	Model     string        `json:"model,omitempty"`
	ToolCalls []interface{} `json:"toolCalls,omitempty"`
}

// geminiTokens holds token counts from a Gemini response message.
type geminiTokens struct {
	Input    int `json:"input"`
	Output   int `json:"output"`
	Cached   int `json:"cached"`
	Thoughts int `json:"thoughts"`
	Tool     int `json:"tool"`
	Total    int `json:"total"`
}

func (g *Gemini) Load() (*ProviderData, error) {
	data := &ProviderData{
		ProviderName: g.Name(),
		Icon:         g.Icon(),
		Color:        g.Color(),
		Metadata:     make(map[string]string),
	}

	sessions, err := g.loadSessions()
	if err != nil {
		return nil, err
	}

	modelAgg := make(map[string]*ModelBreakdown)
	dailyAgg := make(map[string]*DailyUsage)

	for _, sess := range sessions {
		startTime, _ := time.Parse(time.RFC3339, sess.StartTime)
		endTime, _ := time.Parse(time.RFC3339, sess.LastUpdated)
		if endTime.IsZero() {
			endTime = startTime
		}

		var totalInput, totalOutput, totalCached int
		var msgCount, userMsgCount int
		var sessionModel string
		sessionModelTokens := make(map[string]model.TokenUsage)

		for _, msg := range sess.Messages {
			msgCount++
			if msg.Type == "user" {
				userMsgCount++
				continue
			}
			if msg.Type != "gemini" {
				continue
			}
			if msg.Tokens == nil {
				continue
			}

			input := msg.Tokens.Input
			output := msg.Tokens.Output
			cached := msg.Tokens.Cached

			totalInput += input
			totalOutput += output
			totalCached += cached

			m := msg.Model
			if m == "" {
				m = "gemini-2.5-pro"
			}
			sessionModel = m

			existing := sessionModelTokens[m]
			existing.InputTokens += input
			existing.OutputTokens += output
			existing.CacheRead += cached
			sessionModelTokens[m] = existing

			// Aggregate into model breakdown.
			mb, ok := modelAgg[m]
			if !ok {
				mb = &ModelBreakdown{Model: m}
				modelAgg[m] = mb
			}
			mb.InputTokens += input
			mb.OutputTokens += output
			mb.CacheRead += cached
			mb.Generations++
		}

		// Calculate session cost across all models used.
		var sessionCost float64
		totalTokens := totalInput + totalOutput + totalCached
		for m, tu := range sessionModelTokens {
			sessionCost += model.CalculateCost(m, tu)
		}

		si := SessionInfo{
			ID:           sess.SessionID,
			Project:      sess.ProjectHash,
			StartTime:    startTime,
			EndTime:      endTime,
			Messages:     msgCount,
			UserMessages: userMsgCount,
			Tokens:       totalTokens,
			Cost:         sessionCost,
			Model:        sessionModel,
		}
		data.Sessions = append(data.Sessions, si)
		data.TotalCost += sessionCost

		// Track first/last seen.
		if !startTime.IsZero() {
			if data.FirstSeen.IsZero() || startTime.Before(data.FirstSeen) {
				data.FirstSeen = startTime
			}
		}
		if !endTime.IsZero() && endTime.After(data.LastSeen) {
			data.LastSeen = endTime
		}

		// Aggregate daily usage from session start date.
		if !startTime.IsZero() {
			dateKey := startTime.Format("2006-01-02")
			du, ok := dailyAgg[dateKey]
			if !ok {
				du = &DailyUsage{Date: dateKey}
				dailyAgg[dateKey] = du
			}
			du.Cost += sessionCost
			du.Tokens += totalTokens
			du.Messages += msgCount
			du.Sessions++
		}
	}

	// Finalize model breakdowns with cost.
	for m, mb := range modelAgg {
		mb.Cost = model.CalculateCost(m, model.TokenUsage{
			InputTokens:  mb.InputTokens,
			OutputTokens: mb.OutputTokens,
			CacheRead:    mb.CacheRead,
		})
		data.Models = append(data.Models, *mb)
	}
	sort.Slice(data.Models, func(i, j int) bool {
		return data.Models[i].Cost > data.Models[j].Cost
	})

	// Finalize daily usage sorted by date.
	for _, du := range dailyAgg {
		data.DailyUsage = append(data.DailyUsage, *du)
	}
	sort.Slice(data.DailyUsage, func(i, j int) bool {
		return data.DailyUsage[i].Date < data.DailyUsage[j].Date
	})

	return data, nil
}

// loadSessions walks the Gemini tmp directories and parses all session JSON files.
func (g *Gemini) loadSessions() ([]geminiSession, error) {
	tmpDir := filepath.Join(g.ConfigDir, "tmp")
	entries, err := os.ReadDir(tmpDir)
	if err != nil {
		return nil, err
	}

	var sessions []geminiSession
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		chatsDir := filepath.Join(tmpDir, e.Name(), "chats")
		chatFiles, err := os.ReadDir(chatsDir)
		if err != nil {
			continue
		}
		for _, cf := range chatFiles {
			if cf.IsDir() || !strings.HasPrefix(cf.Name(), "session-") || !strings.HasSuffix(cf.Name(), ".json") {
				continue
			}
			sess, err := g.parseSession(filepath.Join(chatsDir, cf.Name()))
			if err != nil {
				continue
			}
			sessions = append(sessions, sess)
		}
	}
	return sessions, nil
}

// parseSession reads and unmarshals a single Gemini session file.
func (g *Gemini) parseSession(path string) (geminiSession, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return geminiSession{}, err
	}
	var sess geminiSession
	if err := json.Unmarshal(raw, &sess); err != nil {
		return geminiSession{}, err
	}
	return sess, nil
}
