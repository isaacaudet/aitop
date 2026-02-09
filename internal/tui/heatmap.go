package tui

import (
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/isaacaudet/clawdtop/internal/provider"
	"github.com/isaacaudet/clawdtop/internal/tui/components"
)

// Heatmap colors from cool to hot.
var heatColors = []lipgloss.Color{
	lipgloss.Color("#313244"), // empty
	lipgloss.Color("#45475a"), // very low
	lipgloss.Color("#585b70"), // low
	lipgloss.Color("#74c7ec"), // medium-low
	lipgloss.Color("#89b4fa"), // medium
	lipgloss.Color("#b4befe"), // medium-high
	lipgloss.Color("#cba6f7"), // high
	lipgloss.Color("#f5c2e7"), // very high
	lipgloss.Color("#f38ba8"), // intense
}

func renderHeatmap(aggData *provider.AggregatedData, width int) string {
	if aggData == nil {
		return StyleError.Render("No data loaded.")
	}

	var sb strings.Builder

	// Build daily data maps for cost and messages.
	costMap := make(map[string]float64)
	msgMap := make(map[string]float64)
	var maxCost, maxMsg float64
	for _, d := range aggData.DailyUsage {
		val := d.Cost
		if val == 0 {
			val = float64(d.Tokens) / 1_000_000
		}
		if val == 0 {
			val = float64(d.Generations)
		}
		costMap[d.Date] = val
		if val > maxCost {
			maxCost = val
		}
		msgVal := float64(d.Messages)
		msgMap[d.Date] = msgVal
		if msgVal > maxMsg {
			maxMsg = msgVal
		}
	}

	// Render cost heatmap (16 weeks).
	sb.WriteString(StyleSectionTitle.Render("Cost Heatmap"))
	sb.WriteString("\n\n")
	sb.WriteString(renderHeatmapGrid(costMap, maxCost, 16, width))
	sb.WriteString("\n")

	// Render messages heatmap (16 weeks).
	sb.WriteString(StyleSectionTitle.Render("Messages Heatmap"))
	sb.WriteString("\n\n")
	sb.WriteString(renderHeatmapGrid(msgMap, maxMsg, 16, width))
	sb.WriteString("\n")

	// Legend.
	sb.WriteString("  Less ")
	for _, c := range heatColors {
		sb.WriteString(lipgloss.NewStyle().Foreground(c).Render("████"))
	}
	sb.WriteString(" More\n\n")

	// Daily cost sparkline (30 days) - full width.
	sb.WriteString(StyleSectionTitle.Render("Daily Cost Trend (30 Days)"))
	sb.WriteString("\n")

	var costVals []float64
	var costDates []string
	last30 := aggData.DailyUsage
	if len(last30) > 30 {
		last30 = last30[len(last30)-30:]
	}
	for _, d := range last30 {
		costVals = append(costVals, d.Cost)
		costDates = append(costDates, d.Date)
	}
	sb.WriteString(WideSparkline(costVals, costDates, ColorGreen, width))
	sb.WriteString("\n")

	// Daily message sparkline.
	sb.WriteString(StyleSectionTitle.Render("Daily Messages (30 Days)"))
	sb.WriteString("\n")
	var msgVals []float64
	var msgDates []string
	for _, d := range last30 {
		msgVals = append(msgVals, float64(d.Messages))
		msgDates = append(msgDates, d.Date)
	}
	sb.WriteString(WideSparkline(msgVals, msgDates, ColorBlue, width))
	sb.WriteString("\n")

	// Summary stats.
	if len(aggData.DailyUsage) > 0 {
		var totalCost float64
		var totalMessages int
		var totalGens int
		var activeDays int
		for _, d := range aggData.DailyUsage {
			totalCost += d.Cost
			totalMessages += d.Messages
			totalGens += d.Generations
			if d.Cost > 0 || d.Messages > 0 || d.Generations > 0 {
				activeDays++
			}
		}

		sb.WriteString(StyleSectionTitle.Render("Summary"))
		sb.WriteString("\n")
		sb.WriteString(fmt.Sprintf("  Active days:    %s\n", StyleStatValue.Render(fmt.Sprintf("%d", activeDays))))
		sb.WriteString(fmt.Sprintf("  Total messages: %s\n", StyleStatValue.Render(components.FormatCount(totalMessages))))
		if totalGens > 0 {
			sb.WriteString(fmt.Sprintf("  Generations:    %s\n", StyleStatValue.Render(components.FormatCount(totalGens))))
		}
		sb.WriteString(fmt.Sprintf("  Total cost:     %s\n", StyleStatCost.Render(fmt.Sprintf("$%.2f", totalCost))))
		if activeDays > 0 {
			sb.WriteString(fmt.Sprintf("  Avg cost/day:   %s\n", StyleStatCost.Render(fmt.Sprintf("$%.2f", totalCost/float64(activeDays)))))
		}
	}

	return sb.String()
}

func renderHeatmapGrid(dailyMap map[string]float64, maxVal float64, weeks int, width int) string {
	var sb strings.Builder

	now := time.Now()
	// Start from N weeks ago, aligned to Sunday.
	start := now.AddDate(0, 0, -(weeks*7 - 1))
	for start.Weekday() != time.Sunday {
		start = start.AddDate(0, 0, -1)
	}

	dayNames := []string{"   ", "Mon", "   ", "Wed", "   ", "Fri", "   "}

	// Header - month labels.
	sb.WriteString("       ")
	for w := 0; w < weeks; w++ {
		weekStart := start.AddDate(0, 0, w*7)
		if weekStart.Day() <= 7 {
			label := weekStart.Format("Jan")
			sb.WriteString(StyleMuted.Render(label))
			padding := 5 - len(label)
			if padding > 0 {
				sb.WriteString(strings.Repeat(" ", padding))
			}
		} else {
			sb.WriteString("     ")
		}
	}
	sb.WriteString("\n")

	// Rows (one per day of week) with double-width cells.
	for d := 0; d < 7; d++ {
		sb.WriteString("  " + StyleMuted.Render(dayNames[d]) + " ")

		for w := 0; w < weeks; w++ {
			date := start.AddDate(0, 0, w*7+d)
			dateStr := date.Format("2006-01-02")
			val := dailyMap[dateStr]

			colorIdx := 0
			if maxVal > 0 && val > 0 {
				// Log scale so low values still show color instead of one spike dominating.
				logVal := math.Log1p(val)
				logMax := math.Log1p(maxVal)
				colorIdx = int(logVal / logMax * float64(len(heatColors)-1))
				if colorIdx < 1 {
					colorIdx = 1
				}
				if colorIdx >= len(heatColors) {
					colorIdx = len(heatColors) - 1
				}
			}

			if date.After(now) {
				sb.WriteString("     ")
			} else {
				style := lipgloss.NewStyle().Foreground(heatColors[colorIdx])
				sb.WriteString(style.Render("████"))
				sb.WriteString(" ")
			}
		}
		sb.WriteString("\n")
	}

	return sb.String()
}
