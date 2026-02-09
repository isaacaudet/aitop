package parser

import (
	"bufio"
	"encoding/json"
	"os"
	"time"

	"github.com/isaacaudet/aitop/internal/model"
)

// ParseSessionFile parses a single JSONL session file into a Session.
func ParseSessionFile(path, project string) (*model.Session, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	session := &model.Session{
		Project: project,
		Models:  make(map[string]model.TokenUsage),
	}

	scanner := bufio.NewScanner(f)
	scanner.Buffer(make([]byte, 1024*1024), 1024*1024) // 1MB line buffer

	for scanner.Scan() {
		var msg model.SessionMessage
		if err := json.Unmarshal(scanner.Bytes(), &msg); err != nil {
			continue // skip malformed lines
		}

		if session.ID == "" && msg.SessionID != "" {
			session.ID = msg.SessionID
		}

		if msg.Timestamp != "" {
			if ts, err := time.Parse(time.RFC3339, msg.Timestamp); err == nil {
				if session.StartTime.IsZero() || ts.Before(session.StartTime) {
					session.StartTime = ts
				}
				if ts.After(session.EndTime) {
					session.EndTime = ts
				}
			}
		}

		if msg.Type == "user" {
			session.UserMessages++
		}

		if msg.Type == "assistant" && msg.Message != nil && msg.Message.Usage != nil {
			session.MessageCount++
			u := msg.Message.Usage
			modelName := msg.Message.Model
			if modelName == "" {
				continue
			}

			session.TokenUsage.InputTokens += u.InputTokens
			session.TokenUsage.OutputTokens += u.OutputTokens
			session.TokenUsage.CacheRead += u.CacheReadInputTokens
			session.TokenUsage.CacheWrite += u.CacheCreationInputTokens

			mu := session.Models[modelName]
			mu.InputTokens += u.InputTokens
			mu.OutputTokens += u.OutputTokens
			mu.CacheRead += u.CacheReadInputTokens
			mu.CacheWrite += u.CacheCreationInputTokens
			session.Models[modelName] = mu
		}
	}

	return session, scanner.Err()
}
