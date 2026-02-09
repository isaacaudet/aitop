package components

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

var (
	headerStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#b4befe")).
			Bold(true)

	rowStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#cdd6f4"))

	altRowStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#a6adc8"))

	selectedStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#1e1e2e")).
			Background(lipgloss.Color("#b4befe")).
			Bold(true)
)

// Column defines a table column.
type Column struct {
	Title string
	Width int
	Align int // 0=left, 1=right
}

// Table renders a styled table.
type Table struct {
	Columns     []Column
	Rows        [][]string
	Selected    int
	ScrollOffset int
	MaxVisible  int
}

// Render produces the table as a string.
func (t Table) Render() string {
	var sb strings.Builder

	// Header.
	var headers []string
	for _, col := range t.Columns {
		cell := padCell(col.Title, col.Width, col.Align)
		headers = append(headers, headerStyle.Render(cell))
	}
	sb.WriteString(strings.Join(headers, " "))
	sb.WriteString("\n")

	// Separator.
	var sep []string
	for _, col := range t.Columns {
		sep = append(sep, headerStyle.Render(strings.Repeat("─", col.Width)))
	}
	sb.WriteString(strings.Join(sep, " "))
	sb.WriteString("\n")

	// Rows.
	maxVisible := t.MaxVisible
	if maxVisible == 0 {
		maxVisible = len(t.Rows)
	}

	start := t.ScrollOffset
	end := start + maxVisible
	if end > len(t.Rows) {
		end = len(t.Rows)
	}

	for i := start; i < end; i++ {
		row := t.Rows[i]
		var cells []string
		for j, col := range t.Columns {
			val := ""
			if j < len(row) {
				val = row[j]
			}
			cell := padCell(val, col.Width, col.Align)

			if i == t.Selected {
				cells = append(cells, selectedStyle.Render(cell))
			} else if i%2 == 0 {
				cells = append(cells, rowStyle.Render(cell))
			} else {
				cells = append(cells, altRowStyle.Render(cell))
			}
		}
		sb.WriteString(strings.Join(cells, " "))
		sb.WriteString("\n")
	}

	// Scroll indicator.
	if len(t.Rows) > maxVisible {
		info := fmt.Sprintf(" %d-%d of %d", start+1, end, len(t.Rows))
		sb.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("#45475a")).Render(info))
		sb.WriteString("\n")
	}

	return sb.String()
}

func padCell(s string, width int, align int) string {
	if len(s) > width {
		s = s[:width]
	}
	if align == 1 { // right-align
		return fmt.Sprintf("%*s", width, s)
	}
	return fmt.Sprintf("%-*s", width, s)
}
