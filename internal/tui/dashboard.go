package tui

import (
	"fmt"
	"sort"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/isaacaudet/clawdtop/internal/config"
	"github.com/isaacaudet/clawdtop/internal/model"
	"github.com/isaacaudet/clawdtop/internal/provider"
	"github.com/isaacaudet/clawdtop/internal/tui/components"
)

func renderDashboard(cache *model.StatsCache, aggData *provider.AggregatedData, width int, cfg config.Config) string {
	if cache == nil {
		return StyleError.Render("No data loaded. Check ~/.claude/stats-cache.json")
	}

	var sb strings.Builder

	today, week, month, allTime := model.ComputeSummaries(cache)

	// Summary boxes row - all same width with sparklines.
	boxWidth := (width - 8) / 4
	if boxWidth < 20 {
		boxWidth = 20
	}

	days := model.AggregateDaily(cache)

	// Today sparkline (hourly activity).
	var todaySparkVals []float64
	for h := 0; h < 24; h++ {
		count := cache.HourCounts[fmt.Sprintf("%d", h)]
		todaySparkVals = append(todaySparkVals, float64(count))
	}
	todaySpark := Sparkline(todaySparkVals, ColorPeach)

	// Week sparkline (last 7 days).
	var weekSparkVals []float64
	recentDays := days
	if len(recentDays) > 7 {
		recentDays = recentDays[len(recentDays)-7:]
	}
	for _, d := range recentDays {
		weekSparkVals = append(weekSparkVals, float64(d.TotalTokens))
	}
	weekSpark := Sparkline(weekSparkVals, ColorBlue)

	// Month sparkline (last 30 days).
	var monthSparkVals []float64
	monthDays := days
	if len(monthDays) > 30 {
		monthDays = monthDays[len(monthDays)-30:]
	}
	for _, d := range monthDays {
		monthSparkVals = append(monthSparkVals, float64(d.TotalTokens))
	}
	monthSpark := Sparkline(monthSparkVals, ColorGreen)

	// All-time sparkline (daily cost over full history, sampled).
	var allTimeSparkVals []float64
	allTimeDays := days
	step := 1
	if len(allTimeDays) > 20 {
		step = len(allTimeDays) / 20
	}
	for i := 0; i < len(allTimeDays); i += step {
		allTimeSparkVals = append(allTimeSparkVals, allTimeDays[i].Cost)
	}
	allTimeSpark := Sparkline(allTimeSparkVals, ColorMauve)

	boxes := lipgloss.JoinHorizontal(lipgloss.Top,
		components.StatBoxWithSparkline(today.Label, today.TotalTokens, today.Cost, today.Messages, boxWidth, todaySpark),
		components.StatBoxWithSparkline(week.Label, week.TotalTokens, week.Cost, week.Messages, boxWidth, weekSpark),
		components.StatBoxWithSparkline(month.Label, month.TotalTokens, month.Cost, month.Messages, boxWidth, monthSpark),
		components.StatBoxWithSparkline(allTime.Label, allTime.TotalTokens, allTime.Cost, allTime.Messages, boxWidth, allTimeSpark),
	)
	sb.WriteString(boxes)
	sb.WriteString("\n\n")

	// Multi-provider summary.
	if aggData != nil && len(aggData.Providers) > 1 {
		sb.WriteString(StyleSectionTitle.Render("AI Tools"))
		sb.WriteString("\n")
		barWidth := width - 50
		if barWidth < 20 {
			barWidth = 20
		}

		var segments []BarSegment
		for _, p := range aggData.Providers {
			iconStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(p.Color))
			costStr := ""
			if p.TotalCost > 0 {
				costStr = StyleStatCost.Render(fmt.Sprintf("$%.2f", p.TotalCost))
			}
			genStr := ""
			if p.Generations > 0 {
				genStr = StyleStatValue.Render(fmt.Sprintf("%d gens", p.Generations))
			}
			sessStr := ""
			if len(p.Sessions) > 0 {
				sessStr = StyleMuted.Render(fmt.Sprintf("%d sessions", len(p.Sessions)))
			}

			info := fmt.Sprintf("  %s  %-14s %s %s %s",
				iconStyle.Render(p.Icon),
				iconStyle.Render(p.ProviderName),
				costStr,
				genStr,
				sessStr,
			)
			sb.WriteString(info)
			sb.WriteString("\n")

			if p.TotalCost > 0 {
				segments = append(segments, BarSegment{
					Label: p.ProviderName,
					Value: p.TotalCost,
					Color: lipgloss.Color(p.Color),
				})
			}
		}

		if len(segments) > 0 {
			sb.WriteString("  ")
			sb.WriteString(StackedBar(segments, barWidth))
			sb.WriteString("\n")
		}
		sb.WriteString("\n")
	}

	// Daily usage bar chart (last 30 days) - taller with date labels.
	chartDays := days
	if len(chartDays) > 30 {
		chartDays = chartDays[len(chartDays)-30:]
	}

	var barData []BarData
	for _, d := range chartDays {
		dayLabel := d.Date[8:] // day of month
		barData = append(barData, BarData{
			Label: dayLabel,
			Value: float64(d.TotalTokens),
			Color: ColorBlue,
		})
	}

	chart := BarChart{
		Title:  "Daily Token Usage (Last 30 Days)",
		Data:   barData,
		Height: 12,
	}
	sb.WriteString(chart.Render())
	sb.WriteString("\n")

	// Model breakdown with aligned horizontal bars.
	sb.WriteString(StyleSectionTitle.Render("Model Cost Breakdown"))
	sb.WriteString("\n")

	type modelEntry struct {
		name  string
		total int
		cost  float64
	}
	var models []modelEntry
	var maxCost float64
	var maxLabelLen int
	for name, mu := range cache.ModelUsage {
		total := mu.InputTokens + mu.OutputTokens + mu.CacheReadInputTokens + mu.CacheCreationInputTokens
		cost := model.CalculateCostFromModelUsage(name, mu)
		displayName := model.NormalizeModelName(name)
		models = append(models, modelEntry{name: displayName, total: total, cost: cost})
		if cost > maxCost {
			maxCost = cost
		}
		if len(displayName) > maxLabelLen {
			maxLabelLen = len(displayName)
		}
	}
	sort.Slice(models, func(i, j int) bool { return models[i].cost > models[j].cost })

	barWidth := width - maxLabelLen - 20
	if barWidth < 20 {
		barWidth = 20
	}
	for i, m := range models {
		if m.cost < 0.01 {
			continue
		}
		color := BarColors[i%len(BarColors)]
		bar := HorizontalBarAligned(m.name, m.cost, maxCost, barWidth, maxLabelLen, color)
		sb.WriteString(bar)
		sb.WriteString(StyleStatCost.Render(fmt.Sprintf("  $%.2f", m.cost)))
		sb.WriteString("\n")
	}
	sb.WriteString("\n")

	// Burn rate panel.
	burnRate := model.ComputeBurnRate(days)
	sb.WriteString(StyleSectionTitle.Render("Burn Rate"))
	sb.WriteString("\n")
	sb.WriteString(fmt.Sprintf("  Daily avg:        %s", StyleStatCost.Render(fmt.Sprintf("$%.2f/day", burnRate.DailyAvg))))
	sb.WriteString(fmt.Sprintf("    Projected month:  %s", StyleStatCost.Render(fmt.Sprintf("$%.2f/mo", burnRate.ProjectedMonth))))

	trend := MiniTrend(burnRate.DailyAvg, burnRate.DailyAvg/(1+burnRate.TrendVsLastWeek/100))
	sb.WriteString(fmt.Sprintf("    Trend: %s", trend))
	sb.WriteString("\n\n")

	// Peak hours bar chart.
	var hourBarData []BarData
	for h := 0; h < 24; h++ {
		count := cache.HourCounts[fmt.Sprintf("%d", h)]
		label := fmt.Sprintf("%02d", h)
		hourBarData = append(hourBarData, BarData{
			Label: label,
			Value: float64(count),
			Color: ColorSapphire,
		})
	}
	hourChart := BarChart{
		Title:  "Activity by Hour",
		Data:   hourBarData,
		Height: 8,
	}
	sb.WriteString(hourChart.Render())
	sb.WriteString("\n")

	return sb.String()
}
