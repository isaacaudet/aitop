package cmd

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/isaacaudet/aitop/internal/config"
	"github.com/isaacaudet/aitop/internal/parser"
	"github.com/isaacaudet/aitop/internal/provider"
	"github.com/isaacaudet/aitop/internal/tui"
	"github.com/spf13/cobra"
	"golang.org/x/term"
)

// AllProviders returns all registered providers.
func AllProviders() []provider.Provider {
	return []provider.Provider{
		provider.NewClaude(),
		provider.NewCursor(),
		provider.NewGemini(),
		provider.NewCodex(),
	}
}

var rootCmd = &cobra.Command{
	Use:   "aitop",
	Short: "Interactive terminal dashboard for AI coding tool usage",
	Long:  "aitop - A beautiful TUI for visualizing AI tool usage, costs, and projections across Claude Code, Cursor, Gemini, Codex, and more.",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg := config.Load()

		statsPath := cfg.StatsCachePath
		if statsPath == "" {
			statsPath = parser.DefaultStatsCachePath()
		}

		cache, err := parser.ParseStatsCache(statsPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: Could not load Claude stats cache: %v\n", err)
		}

		providers := AllProviders()

		// Detect available providers.
		var available []string
		for _, p := range providers {
			if p.Available() {
				available = append(available, p.Name())
			}
		}
		if len(available) == 0 && cache == nil {
			return fmt.Errorf("no AI tool data found. Install and use Claude Code, Cursor, or other supported tools")
		}

		if !term.IsTerminal(int(os.Stdin.Fd())) {
			fmt.Fprintf(os.Stderr, "No interactive terminal detected. Falling back to summary mode.\n")
			fmt.Fprintf(os.Stderr, "Run `aitop` from a real terminal for the interactive TUI.\n\n")
			return summaryCmd.RunE(cmd, args)
		}

		m := tui.New(cache, providers)
		p := tea.NewProgram(m, tea.WithAltScreen())
		if _, err := p.Run(); err != nil {
			return err
		}
		return nil
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
