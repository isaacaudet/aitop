package components

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
)

var (
	boxBorder = lipgloss.RoundedBorder()

	boxStyle = lipgloss.NewStyle().
			Border(boxBorder).
			BorderForeground(lipgloss.Color("#45475a")).
			Padding(0, 1)

	labelStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#a6adc8")).
			Bold(true)

	valueStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#cdd6f4")).
			Bold(true)

	costStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#a6e3a1")).
			Bold(true)

	dimStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#45475a"))
)

// StatBox renders a summary stat box (always includes sparkline slot for uniform height).
func StatBox(title string, tokens int, cost float64, messages int, width int) string {
	return StatBoxWithSparkline(title, tokens, cost, messages, width, "")
}

// StatBoxWithSparkline renders a stat box with an inline sparkline.
func StatBoxWithSparkline(title string, tokens int, cost float64, messages int, width int, sparkline string) string {
	content := labelStyle.Render(title) + "\n" +
		costStyle.Render(fmt.Sprintf("$%.2f", cost)) + "\n" +
		valueStyle.Render(FormatTokens(tokens)) + " tokens\n"

	if sparkline != "" {
		content += sparkline + "\n"
	}

	content += dimStyle.Render(fmt.Sprintf("%s msgs", FormatCount(messages)))

	return boxStyle.Width(width).Render(content)
}

// ProviderStatBox renders a stat box for a provider.
func ProviderStatBox(icon, name string, cost float64, generations int, color lipgloss.Color, width int) string {
	iconStyle := lipgloss.NewStyle().Foreground(color).Bold(true)
	nameStyle := lipgloss.NewStyle().Foreground(color)

	var lines []string
	lines = append(lines, iconStyle.Render(icon)+" "+nameStyle.Render(name))
	if cost > 0 {
		lines = append(lines, costStyle.Render(fmt.Sprintf("$%.2f", cost)))
	}
	if generations > 0 {
		lines = append(lines, valueStyle.Render(fmt.Sprintf("%s generations", FormatCount(generations))))
	}

	content := ""
	for i, l := range lines {
		if i > 0 {
			content += "\n"
		}
		content += l
	}

	return boxStyle.Width(width).BorderForeground(color).Render(content)
}

func FormatTokens(n int) string {
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

func FormatCount(n int) string {
	switch {
	case n >= 1_000_000:
		return fmt.Sprintf("%.1fM", float64(n)/1_000_000)
	case n >= 1_000:
		return fmt.Sprintf("%.1fK", float64(n)/1_000)
	default:
		return fmt.Sprintf("%d", n)
	}
}
