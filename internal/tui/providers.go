package tui

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/isaacaudet/aitop/internal/provider"
	"github.com/isaacaudet/aitop/internal/tui/components"
)

func renderProvidersView(aggData *provider.AggregatedData, width int, period timePeriod) string {
	if aggData == nil || len(aggData.Providers) == 0 {
		return StyleError.Render("No providers loaded.")
	}

	var sb strings.Builder

	// Provider cards row.
	boxWidth := (width - 4) / len(aggData.Providers)
	if boxWidth < 24 {
		boxWidth = 24
	}
	if boxWidth > 40 {
		boxWidth = 40
	}

	var cards []string
	for _, p := range aggData.Providers {
		cards = append(cards, components.ProviderStatBox(
			p.Icon, p.ProviderName, p.TotalCost, p.Generations,
			lipgloss.Color(p.Color), boxWidth,
		))
	}
	sb.WriteString(lipgloss.JoinHorizontal(lipgloss.Top, cards...))
	sb.WriteString("\n\n")

	// Determine time filter.
	var filterDate string
	now := time.Now()
	switch period {
	case periodToday:
		filterDate = now.Format("2006-01-02")
	case periodThisWeek:
		filterDate = now.AddDate(0, 0, -7).Format("2006-01-02")
	case periodThisMonth:
		filterDate = now.AddDate(0, -1, 0).Format("2006-01-02")
	default:
		filterDate = ""
	}

	// Per-provider detailed breakdown.
	for _, p := range aggData.Providers {
		provColor := lipgloss.Color(p.Color)
		provStyle := lipgloss.NewStyle().Foreground(provColor).Bold(true)

		sb.WriteString(provStyle.Render(fmt.Sprintf("%s %s", p.Icon, p.ProviderName)))
		sb.WriteString("\n")

		// Filter models by time period if applicable.
		models := p.Models
		if filterDate != "" && len(p.DailyUsage) > 0 {
			// Show filtered cost from daily usage.
			var filteredCost float64
			var filteredTokens int
			for _, d := range p.DailyUsage {
				if d.Date >= filterDate {
					filteredCost += d.Cost
					filteredTokens += d.Tokens
				}
			}
			sb.WriteString(fmt.Sprintf("  Period cost: %s  Tokens: %s\n",
				StyleStatCost.Render(fmt.Sprintf("$%.2f", filteredCost)),
				StyleStatValue.Render(components.FormatTokens(filteredTokens)),
			))
		}

		if len(models) > 0 && p.TotalCost > 0 {
			// Models with cost data.
			sort.Slice(models, func(i, j int) bool { return models[i].Cost > models[j].Cost })

			columns := []components.Column{
				{Title: "Model", Width: 22},
				{Title: "Input", Width: 12, Align: 1},
				{Title: "Output", Width: 12, Align: 1},
				{Title: "Cache Read", Width: 12, Align: 1},
				{Title: "Cache Write", Width: 12, Align: 1},
				{Title: "Cost", Width: 12, Align: 1},
			}

			var rows [][]string
			for _, m := range models {
				rows = append(rows, []string{
					m.Model,
					components.FormatTokens(m.InputTokens),
					components.FormatTokens(m.OutputTokens),
					components.FormatTokens(m.CacheRead),
					components.FormatTokens(m.CacheWrite),
					fmt.Sprintf("$%.2f", m.Cost),
				})
			}

			table := components.Table{
				Columns:    columns,
				Rows:       rows,
				Selected:   -1,
				MaxVisible: len(rows),
			}
			sb.WriteString(table.Render())

		} else if len(models) > 0 {
			// Models without cost (e.g., Cursor file extensions).
			sort.Slice(models, func(i, j int) bool { return models[i].Generations > models[j].Generations })

			barWidth := width - 50
			if barWidth < 20 {
				barWidth = 20
			}

			maxGen := 0
			maxLabelLen := 0
			for _, m := range models {
				if m.Generations > maxGen {
					maxGen = m.Generations
				}
				if len(m.Model) > maxLabelLen {
					maxLabelLen = len(m.Model)
				}
			}
			if maxLabelLen > 20 {
				maxLabelLen = 20
			}

			shown := models
			if len(shown) > 10 {
				shown = shown[:10]
			}
			alignedBarWidth := width - maxLabelLen - 15
			if alignedBarWidth < 10 {
				alignedBarWidth = 10
			}
			for i, m := range shown {
				color := BarColors[i%len(BarColors)]
				bar := HorizontalBarAligned(m.Model, float64(m.Generations), float64(maxGen), alignedBarWidth, maxLabelLen, color)
				sb.WriteString(bar)
				sb.WriteString(StyleStatValue.Render(fmt.Sprintf("  %d", m.Generations)))
				sb.WriteString("\n")
			}
		} else if len(p.Sessions) > 0 {
			// Provider with sessions but no model breakdown (show token summary).
			var totalTokens int
			for _, s := range p.Sessions {
				totalTokens += s.Tokens
			}
			sb.WriteString(fmt.Sprintf("  %s sessions  %s tokens\n",
				StyleStatValue.Render(fmt.Sprintf("%d", len(p.Sessions))),
				StyleStatValue.Render(components.FormatTokens(totalTokens)),
			))
		}

		// Cost attribution bar — fixed-width label so all providers align.
		if p.TotalCost > 0 {
			var segments []BarSegment
			for i, m := range models {
				if m.Cost > 0 {
					segments = append(segments, BarSegment{
						Label: m.Model,
						Value: m.Cost,
						Color: BarColors[i%len(BarColors)],
					})
				}
			}
			if len(segments) > 0 {
				costLabel := "  Cost:              "
				sb.WriteString("\n" + StyleMuted.Render(costLabel))
				sb.WriteString(StackedBar(segments, width-len(costLabel)-2))
				sb.WriteString("\n")
			}
		}

		sb.WriteString("\n")
	}

	// Grand total.
	sb.WriteString(StyleSectionTitle.Render("Grand Total"))
	sb.WriteString("\n")
	sb.WriteString(fmt.Sprintf("  %s %s\n",
		StyleStatLabel.Render("Total Cost Across All Providers:"),
		StyleStatCost.Render(fmt.Sprintf("$%.2f", aggData.TotalCost)),
	))

	return sb.String()
}
