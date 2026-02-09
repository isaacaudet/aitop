package tui

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/isaacaudet/clawdtop/internal/provider"
)

const (
	liveBucketHours = 6    // Each bucket = 6 hours
	liveDaysDefault = 7    // Show last 7 days
	liveDaysMax     = 30   // Max 30 days of history
)

type liveView struct {
	scrollOffset int // bucket offset from right edge
	daysWindow   int
}

func newLiveView() liveView {
	return liveView{
		daysWindow: liveDaysDefault,
	}
}

func (lv *liveView) scrollLeft() {
	lv.scrollOffset++
	max := (liveDaysMax * 24 / liveBucketHours) - (lv.daysWindow * 24 / liveBucketHours)
	if lv.scrollOffset > max {
		lv.scrollOffset = max
	}
}

func (lv *liveView) scrollRight() {
	lv.scrollOffset--
	if lv.scrollOffset < 0 {
		lv.scrollOffset = 0
	}
}

func (lv liveView) render(aggData *provider.AggregatedData, width int) string {
	if aggData == nil || len(aggData.Providers) == 0 {
		return StyleMuted.Render("  No provider data available.")
	}

	var sb strings.Builder
	sb.WriteString(StyleSectionTitle.Render("Activity Timeline"))
	sb.WriteString("\n\n")

	now := time.Now()
	bucketsPerWindow := lv.daysWindow * 24 / liveBucketHours
	maxBuckets := width - 22
	if bucketsPerWindow > maxBuckets {
		bucketsPerWindow = maxBuckets
	}

	// Build time buckets for each provider using session start times.
	type providerRow struct {
		name    string
		icon    string
		color   lipgloss.Color
		buckets []float64
	}

	var rows []providerRow
	var globalMax float64

	for _, p := range aggData.Providers {
		if len(p.DailyUsage) == 0 && len(p.Sessions) == 0 {
			continue
		}

		buckets := make([]float64, bucketsPerWindow)

		// Use session data for granular bucketing.
		for _, s := range p.Sessions {
			hoursAgo := now.Sub(s.StartTime).Hours() - float64(lv.scrollOffset*liveBucketHours)
			bucketIdx := bucketsPerWindow - 1 - int(hoursAgo/float64(liveBucketHours))
			if bucketIdx >= 0 && bucketIdx < bucketsPerWindow {
				val := float64(s.Tokens)
				if val == 0 {
					val = s.Cost * 10000
				}
				if val == 0 {
					val = 1
				}
				buckets[bucketIdx] += val
				if buckets[bucketIdx] > globalMax {
					globalMax = buckets[bucketIdx]
				}
			}
		}

		// Also fill from daily usage for providers without session-level times.
		for _, d := range p.DailyUsage {
			dt, err := time.Parse("2006-01-02", d.Date)
			if err != nil {
				continue
			}
			// Center on noon for daily data.
			dt = dt.Add(12 * time.Hour)
			hoursAgo := now.Sub(dt).Hours() - float64(lv.scrollOffset*liveBucketHours)
			bucketIdx := bucketsPerWindow - 1 - int(hoursAgo/float64(liveBucketHours))
			if bucketIdx >= 0 && bucketIdx < bucketsPerWindow {
				val := float64(d.Tokens)
				if val == 0 {
					val = d.Cost * 10000
				}
				if val == 0 {
					val = float64(d.Generations)
				}
				// Only add if no session data already filled this bucket.
				if buckets[bucketIdx] == 0 && val > 0 {
					buckets[bucketIdx] = val
					if val > globalMax {
						globalMax = val
					}
				}
			}
		}

		rows = append(rows, providerRow{
			name:    p.ProviderName,
			icon:    p.Icon,
			color:   lipgloss.Color(p.Color),
			buckets: buckets,
		})
	}

	if globalMax == 0 {
		globalMax = 1
	}

	// Intensity characters — 3 rows tall per provider for better visibility.
	intensityBlocks := []string{" ", "░", "▒", "▓", "█"}

	maxNameLen := 0
	for _, r := range rows {
		if len(r.name) > maxNameLen {
			maxNameLen = len(r.name)
		}
	}

	for _, r := range rows {
		style := lipgloss.NewStyle().Foreground(r.color)
		label := fmt.Sprintf("  %s %-*s ", r.icon, maxNameLen, r.name)
		sb.WriteString(style.Render(label))

		for _, val := range r.buckets {
			intensityIdx := int(val / globalMax * float64(len(intensityBlocks)-1))
			if intensityIdx < 0 {
				intensityIdx = 0
			}
			if intensityIdx >= len(intensityBlocks) {
				intensityIdx = len(intensityBlocks) - 1
			}
			sb.WriteString(style.Render(intensityBlocks[intensityIdx]))
		}
		sb.WriteString("\n")
	}

	// Time axis.
	labelPad := strings.Repeat(" ", maxNameLen+5)
	sb.WriteString(labelPad)
	bucketsPerDay := 24 / liveBucketHours
	for i := 0; i < bucketsPerWindow; i++ {
		if i%bucketsPerDay == 0 {
			hoursFromEnd := (bucketsPerWindow - 1 - i) * liveBucketHours
			hoursFromEnd += lv.scrollOffset * liveBucketHours
			t := now.Add(-time.Duration(hoursFromEnd) * time.Hour)
			label := t.Format("Jan02")
			if len(label) <= bucketsPerDay || i+len(label) <= bucketsPerWindow {
				sb.WriteString(StyleMuted.Render(label))
				i += len(label) - 1
			} else {
				sb.WriteString(" ")
			}
		} else {
			sb.WriteString(" ")
		}
	}
	sb.WriteString("\n\n")

	// Legend.
	sb.WriteString("  Intensity: ")
	for _, ch := range intensityBlocks {
		if ch == " " {
			sb.WriteString(StyleMuted.Render("·"))
		} else {
			sb.WriteString(StyleMuted.Render(ch))
		}
		sb.WriteString(" ")
	}
	sb.WriteString(StyleMuted.Render(fmt.Sprintf(" (max: %s tokens per %dh bucket)", formatAxisValue(globalMax), liveBucketHours)))
	sb.WriteString("\n\n")

	if lv.scrollOffset > 0 {
		sb.WriteString(StyleMuted.Render(fmt.Sprintf("  Showing %d days ago. Press l to return to now, h to scroll back.", lv.scrollOffset*liveBucketHours/24)))
		sb.WriteString("\n")
	} else {
		sb.WriteString(StyleMuted.Render("  Auto-refreshes every 5s. Press h to scroll back in time."))
		sb.WriteString("\n")
	}

	return sb.String()
}
