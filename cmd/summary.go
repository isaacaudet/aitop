package cmd

import (
	"fmt"

	"github.com/isaacaudet/clawdtop/internal/config"
	"github.com/isaacaudet/clawdtop/internal/model"
	"github.com/isaacaudet/clawdtop/internal/parser"
	"github.com/isaacaudet/clawdtop/internal/provider"
	"github.com/spf13/cobra"
)

var summaryCmd = &cobra.Command{
	Use:   "summary",
	Short: "Print a one-shot usage summary across all AI tools",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg := config.Load()

		statsPath := cfg.StatsCachePath
		if statsPath == "" {
			statsPath = parser.DefaultStatsCachePath()
		}

		// Load all providers.
		providers := AllProviders()
		aggData := provider.LoadAll(providers)

		fmt.Println("clawdtop — AI Usage Dashboard")
		fmt.Println("═══════════════════════════════════════════════════════")
		fmt.Println()

		// Per-provider summary.
		for _, p := range aggData.Providers {
			fmt.Printf("  %s %s", p.Icon, p.ProviderName)
			if p.TotalCost > 0 {
				fmt.Printf("  $%.2f", p.TotalCost)
			}
			if p.Generations > 0 {
				fmt.Printf("  %d generations", p.Generations)
			}
			if len(p.Sessions) > 0 {
				fmt.Printf("  %d sessions", len(p.Sessions))
			}
			var totalTokens int
			for _, m := range p.Models {
				totalTokens += m.InputTokens + m.OutputTokens + m.CacheRead + m.CacheWrite
			}
			if totalTokens > 0 {
				fmt.Printf("  %s tokens", formatTokens(totalTokens))
			}
			fmt.Println()
		}
		fmt.Println()

		// Claude-specific detailed stats.
		cache, err := parser.ParseStatsCache(statsPath)
		if err == nil {
			today, week, month, allTime := model.ComputeSummaries(cache)
			days := model.AggregateDaily(cache)
			burn := model.ComputeBurnRate(days)

			fmt.Println("Claude Code Detailed Stats")
			fmt.Println("─────────────────────────────────────────────────────")

			printPeriod := func(p model.PeriodSummary) {
				fmt.Printf("  %-14s  $%8.2f  %8s tokens  %5d msgs  %4d sessions\n",
					p.Label, p.Cost, formatTokens(p.TotalTokens), p.Messages, p.Sessions)
			}
			printPeriod(today)
			printPeriod(week)
			printPeriod(month)
			printPeriod(allTime)

			fmt.Println()
			fmt.Println("  Burn Rate")
			fmt.Printf("    Daily average:    $%.2f/day\n", burn.DailyAvg)
			fmt.Printf("    Projected month:  $%.2f/mo\n", burn.ProjectedMonth)
			fmt.Printf("    Trend vs last wk: %+.1f%%\n", burn.TrendVsLastWeek)

			fmt.Println()
			fmt.Println("  Model Breakdown")
			totalCost := model.TotalCostFromModelUsage(cache.ModelUsage)
			for name, mu := range cache.ModelUsage {
				cost := model.CalculateCostFromModelUsage(name, mu)
				if cost < 0.01 {
					continue
				}
				pct := float64(0)
				if totalCost > 0 {
					pct = cost / totalCost * 100
				}
				fmt.Printf("    %-28s  $%8.2f  (%5.1f%%)\n",
					model.NormalizeModelName(name), cost, pct)
			}
			fmt.Println()
			fmt.Printf("  %d sessions, %d messages since %s\n",
				cache.TotalSessions, cache.TotalMessages, cache.FirstSessionDate[:10])
		}

		fmt.Println()
		fmt.Println("═══════════════════════════════════════════════════════")
		fmt.Printf("  Grand Total: $%.2f across %d providers\n", aggData.TotalCost, len(aggData.Providers))

		return nil
	},
}

func formatTokens(n int) string {
	switch {
	case n >= 1_000_000_000:
		return fmt.Sprintf("%.1fB", float64(n)/1_000_000_000)
	case n >= 1_000_000:
		return fmt.Sprintf("%.1fM", float64(n)/1_000_000)
	case n >= 1_000:
		return fmt.Sprintf("%.1fK", float64(n)/1_000)
	default:
		return fmt.Sprintf("%d", n)
	}
}

func init() {
	rootCmd.AddCommand(summaryCmd)
}
