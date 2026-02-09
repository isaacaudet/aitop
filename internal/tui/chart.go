package tui

import (
	"fmt"
	"math"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// BarChart renders a Unicode vertical bar chart.
type BarChart struct {
	Title  string
	Data   []BarData
	Width  int
	Height int
}

type BarData struct {
	Label string
	Value float64
	Color lipgloss.Color
}

func (bc BarChart) Render() string {
	if len(bc.Data) == 0 {
		return StyleMuted.Render("  No data")
	}

	height := bc.Height
	if height == 0 {
		height = 12
	}

	var maxVal float64
	for _, d := range bc.Data {
		if d.Value > maxVal {
			maxVal = d.Value
		}
	}
	if maxVal == 0 {
		maxVal = 1
	}

	blocks := []string{"▁", "▂", "▃", "▄", "▅", "▆", "▇", "█"}

	var sb strings.Builder
	if bc.Title != "" {
		sb.WriteString(StyleSectionTitle.Render(bc.Title))
		sb.WriteString("\n")
	}

	// Render bars row by row with Y-axis labels.
	yLabelWidth := len(formatAxisValue(maxVal)) + 1
	for row := height; row >= 1; row-- {
		// Y-axis label on top and bottom rows.
		if row == height {
			label := formatAxisValue(maxVal)
			sb.WriteString(StyleMuted.Render(fmt.Sprintf("%*s ", yLabelWidth, label)))
		} else if row == 1 {
			sb.WriteString(StyleMuted.Render(fmt.Sprintf("%*s ", yLabelWidth, "0")))
		} else {
			sb.WriteString(strings.Repeat(" ", yLabelWidth+1))
		}

		threshold := float64(row) / float64(height) * maxVal
		prevThreshold := float64(row-1) / float64(height) * maxVal

		for i, d := range bc.Data {
			color := d.Color
			if color == "" {
				color = BarColors[i%len(BarColors)]
			}
			style := lipgloss.NewStyle().Foreground(color)

			if d.Value >= threshold {
				sb.WriteString(style.Render("█"))
			} else if d.Value > prevThreshold {
				frac := (d.Value - prevThreshold) / (threshold - prevThreshold)
				idx := int(frac * float64(len(blocks)-1))
				if idx < 0 {
					idx = 0
				}
				if idx >= len(blocks) {
					idx = len(blocks) - 1
				}
				sb.WriteString(style.Render(blocks[idx]))
			} else {
				sb.WriteString(" ")
			}
		}
		sb.WriteString("\n")
	}

	// X-axis date labels (every 3rd label).
	sb.WriteString(strings.Repeat(" ", yLabelWidth+1))
	for i, d := range bc.Data {
		if i%3 == 0 && len(d.Label) >= 2 {
			sb.WriteString(StyleMuted.Render(d.Label[:2]))
			// Skip next chars that would overlap
			if i+1 < len(bc.Data) && i+2 < len(bc.Data) {
				// We printed 2 chars, skip 1 position since loop advances
			}
		} else if i%3 == 1 {
			// Space for alignment after a 2-char label
		} else {
			sb.WriteString(" ")
		}
	}
	sb.WriteString("\n")

	return sb.String()
}

// HorizontalBar renders a single horizontal bar with label and value.
// labelWidth controls fixed-width label padding for alignment. Pass 0 for default (18).
func HorizontalBar(label string, value, maxVal float64, width int, color lipgloss.Color) string {
	return HorizontalBarAligned(label, value, maxVal, width, 0, color)
}

// HorizontalBarAligned renders a horizontal bar with explicit label width for alignment.
// Pads the raw label BEFORE styling to ensure visual alignment (ANSI codes don't affect padding).
func HorizontalBarAligned(label string, value, maxVal float64, width, labelWidth int, color lipgloss.Color) string {
	if maxVal == 0 {
		maxVal = 1
	}
	if labelWidth == 0 {
		labelWidth = 18
	}
	barWidth := int(value / maxVal * float64(width))
	if barWidth < 0 {
		barWidth = 0
	}
	if barWidth > width {
		barWidth = width
	}

	// Pad the label to fixed width BEFORE applying style, so ANSI escapes don't break alignment.
	paddedLabel := label
	if len(paddedLabel) > labelWidth {
		paddedLabel = paddedLabel[:labelWidth]
	}
	for len(paddedLabel) < labelWidth {
		paddedLabel += " "
	}

	style := lipgloss.NewStyle().Foreground(color)
	bar := style.Render(strings.Repeat("█", barWidth))
	if width-barWidth > 0 {
		bar += lipgloss.NewStyle().Foreground(ColorOverlay).Render(strings.Repeat("░", width-barWidth))
	}

	return fmt.Sprintf("  %s %s", StyleSubtitle.Render(paddedLabel), bar)
}

// Sparkline renders inline sparkline from values.
func Sparkline(values []float64, color lipgloss.Color) string {
	if len(values) == 0 {
		return ""
	}

	blocks := []rune{'▁', '▂', '▃', '▄', '▅', '▆', '▇', '█'}

	var maxVal float64
	for _, v := range values {
		if v > maxVal {
			maxVal = v
		}
	}
	if maxVal == 0 {
		maxVal = 1
	}

	style := lipgloss.NewStyle().Foreground(color)
	var sb strings.Builder
	for _, v := range values {
		idx := int(v / maxVal * float64(len(blocks)-1))
		if idx < 0 {
			idx = 0
		}
		if idx >= len(blocks) {
			idx = len(blocks) - 1
		}
		sb.WriteRune(blocks[idx])
	}

	return style.Render(sb.String())
}

// WideSparkline renders a full-width sparkline with date labels underneath.
func WideSparkline(values []float64, dates []string, color lipgloss.Color, width int) string {
	if len(values) == 0 {
		return ""
	}

	var sb strings.Builder
	sb.WriteString("  ")
	sb.WriteString(Sparkline(values, color))
	sb.WriteString("\n  ")

	// Show date labels at start, middle, and end.
	if len(dates) > 0 {
		sparkLen := len(values)
		for i := 0; i < sparkLen; i++ {
			if i == 0 || i == sparkLen/2 || i == sparkLen-1 {
				if i < len(dates) {
					label := dates[i]
					if len(label) > 5 {
						label = label[5:] // MM-DD format
					}
					sb.WriteString(StyleMuted.Render(label))
					remaining := sparkLen/2 - len(label)
					if remaining > 0 && i != sparkLen-1 {
						sb.WriteString(strings.Repeat(" ", remaining))
					}
				}
			}
		}
	}
	sb.WriteString("\n")

	return sb.String()
}

// MiniTrend renders a small trend indicator with arrow.
func MiniTrend(current, previous float64) string {
	if previous == 0 {
		return StyleMuted.Render("--")
	}
	change := ((current - previous) / previous) * 100

	var arrow string
	var style lipgloss.Style
	switch {
	case change > 10:
		arrow = "▲"
		style = StyleNegative // spending up = bad
	case change > 0:
		arrow = "△"
		style = StyleWarning
	case change > -10:
		arrow = "▽"
		style = StylePositive // spending down = good
	default:
		arrow = "▼"
		style = StylePositive
	}

	return style.Render(fmt.Sprintf("%s%.0f%%", arrow, math.Abs(change)))
}

// StackedBar renders a stacked horizontal bar with multiple segments.
func StackedBar(segments []BarSegment, width int) string {
	var total float64
	for _, s := range segments {
		total += s.Value
	}
	if total == 0 {
		return strings.Repeat("░", width)
	}

	var sb strings.Builder
	remaining := width
	for i, s := range segments {
		segWidth := int(s.Value / total * float64(width))
		if i == len(segments)-1 {
			segWidth = remaining // Use remaining space for last segment
		}
		if segWidth <= 0 {
			continue
		}
		remaining -= segWidth
		style := lipgloss.NewStyle().Foreground(s.Color)
		sb.WriteString(style.Render(strings.Repeat("█", segWidth)))
	}

	return sb.String()
}

type BarSegment struct {
	Label string
	Value float64
	Color lipgloss.Color
}

// formatAxisValue formats a number for Y-axis display.
func formatAxisValue(v float64) string {
	switch {
	case v >= 1_000_000_000:
		return fmt.Sprintf("%.0fB", v/1_000_000_000)
	case v >= 1_000_000:
		return fmt.Sprintf("%.0fM", v/1_000_000)
	case v >= 1_000:
		return fmt.Sprintf("%.0fK", v/1_000)
	default:
		return fmt.Sprintf("%.0f", v)
	}
}
