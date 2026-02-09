package provider

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

// Cursor implements Provider for Cursor IDE.
type Cursor struct {
	DBPath string
}

func NewCursor() *Cursor {
	home, _ := os.UserHomeDir()
	return &Cursor{
		DBPath: filepath.Join(home, ".cursor", "ai-tracking", "ai-code-tracking.db"),
	}
}

func (c *Cursor) Name() string  { return "Cursor" }
func (c *Cursor) Icon() string  { return "⌘" }
func (c *Cursor) Color() string { return "#f9e2af" } // Yellow

func (c *Cursor) Available() bool {
	_, err := os.Stat(c.DBPath)
	return err == nil
}

func (c *Cursor) Load() (*ProviderData, error) {
	db, err := sql.Open("sqlite3", c.DBPath+"?mode=ro")
	if err != nil {
		return nil, fmt.Errorf("opening cursor db: %w", err)
	}
	defer db.Close()

	data := &ProviderData{
		ProviderName: c.Name(),
		Icon:         c.Icon(),
		Color:        c.Color(),
		Metadata:     make(map[string]string),
	}

	// Total code generations.
	var totalGens int
	err = db.QueryRow("SELECT count(*) FROM ai_code_hashes").Scan(&totalGens)
	if err != nil {
		return nil, err
	}
	data.Generations = totalGens

	// Daily code generations.
	rows, err := db.Query(`
		SELECT date(createdAt/1000, 'unixepoch') as day, count(*) as cnt
		FROM ai_code_hashes
		GROUP BY day
		ORDER BY day
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var day string
		var cnt int
		if err := rows.Scan(&day, &cnt); err != nil {
			continue
		}
		data.DailyUsage = append(data.DailyUsage, DailyUsage{
			Date:        day,
			Generations: cnt,
		})
	}

	// Generations by file extension (as "model" breakdown).
	extRows, err := db.Query(`
		SELECT COALESCE(fileExtension, 'unknown') as ext, count(*) as cnt
		FROM ai_code_hashes
		GROUP BY ext
		ORDER BY cnt DESC
	`)
	if err != nil {
		return nil, err
	}
	defer extRows.Close()

	for extRows.Next() {
		var ext string
		var cnt int
		if err := extRows.Scan(&ext, &cnt); err != nil {
			continue
		}
		data.Models = append(data.Models, ModelBreakdown{
			Model:       ext,
			Generations: cnt,
		})
	}

	// Conversation summaries as sessions.
	sessRows, err := db.Query(`
		SELECT conversationId, COALESCE(title, ''), COALESCE(model, ''),
		       COALESCE(mode, ''), updatedAt
		FROM conversation_summaries
		ORDER BY updatedAt DESC
	`)
	if err != nil {
		return nil, err
	}
	defer sessRows.Close()

	for sessRows.Next() {
		var id, title, model, mode string
		var updatedAt int64
		if err := sessRows.Scan(&id, &title, &model, &mode, &updatedAt); err != nil {
			continue
		}
		t := time.UnixMilli(updatedAt)
		project := title
		if project == "" {
			project = mode
		}
		data.Sessions = append(data.Sessions, SessionInfo{
			ID:        id,
			Project:   project,
			StartTime: t,
			EndTime:   t,
			Model:     model,
		})
	}

	// Date range.
	var minTs, maxTs sql.NullInt64
	db.QueryRow("SELECT min(createdAt), max(createdAt) FROM ai_code_hashes").Scan(&minTs, &maxTs)
	if minTs.Valid {
		data.FirstSeen = time.UnixMilli(minTs.Int64)
	}
	if maxTs.Valid {
		data.LastSeen = time.UnixMilli(maxTs.Int64)
	}

	// Source breakdown metadata.
	sourceRows, err := db.Query("SELECT source, count(*) FROM ai_code_hashes GROUP BY source")
	if err == nil {
		defer sourceRows.Close()
		for sourceRows.Next() {
			var source string
			var cnt int
			if err := sourceRows.Scan(&source, &cnt); err == nil {
				data.Metadata[fmt.Sprintf("source_%s", source)] = fmt.Sprintf("%d", cnt)
			}
		}
	}

	// Note: Cursor doesn't expose token counts or costs locally.
	// Cost estimation is not possible without API access.
	data.Metadata["note"] = "Cursor tracks code generations, not token usage. Cost requires Cursor billing API."

	return data, nil
}
