package parser

import (
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/isaacaudet/clawdtop/internal/model"
)

// DefaultProjectsDir returns the default path to the projects directory.
func DefaultProjectsDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(home, ".claude", "projects")
}

// DecodeProjectPath converts a directory name like "-Users-isaacaudet-myproject" to a readable path.
func DecodeProjectPath(dirName string) string {
	// Replace leading "-" with "/" and subsequent "-" with "/"
	if strings.HasPrefix(dirName, "-") {
		return "/" + strings.ReplaceAll(dirName[1:], "-", "/")
	}
	return dirName
}

// ProjectName extracts a short project name from a decoded path.
func ProjectName(decodedPath string) string {
	parts := strings.Split(strings.TrimRight(decodedPath, "/"), "/")
	if len(parts) == 0 {
		return decodedPath
	}
	// Return last non-empty component.
	for i := len(parts) - 1; i >= 0; i-- {
		if parts[i] != "" {
			return parts[i]
		}
	}
	return decodedPath
}

// LoadAllSessions walks the projects directory and parses all session files.
// Uses goroutines for parallel parsing.
func LoadAllSessions(projectsDir string) ([]*model.Session, error) {
	entries, err := os.ReadDir(projectsDir)
	if err != nil {
		return nil, err
	}

	var (
		mu       sync.Mutex
		wg       sync.WaitGroup
		sessions []*model.Session
		sem      = make(chan struct{}, 8) // limit concurrency
	)

	for _, projEntry := range entries {
		if !projEntry.IsDir() {
			continue
		}

		projDir := filepath.Join(projectsDir, projEntry.Name())
		projName := ProjectName(DecodeProjectPath(projEntry.Name()))

		// Find all JSONL files in this project directory.
		files, err := filepath.Glob(filepath.Join(projDir, "*.jsonl"))
		if err != nil {
			continue
		}

		for _, file := range files {
			wg.Add(1)
			sem <- struct{}{}
			go func(f, proj string) {
				defer wg.Done()
				defer func() { <-sem }()

				session, err := ParseSessionFile(f, proj)
				if err != nil || session.ID == "" {
					return
				}

				mu.Lock()
				sessions = append(sessions, session)
				mu.Unlock()
			}(file, projName)
		}
	}

	wg.Wait()
	return sessions, nil
}
