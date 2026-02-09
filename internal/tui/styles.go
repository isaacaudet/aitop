package tui

import "github.com/charmbracelet/lipgloss"

// Catppuccin Mocha-inspired color palette.
var (
	ColorBase     = lipgloss.Color("#1e1e2e")
	ColorSurface  = lipgloss.Color("#313244")
	ColorOverlay  = lipgloss.Color("#45475a")
	ColorText     = lipgloss.Color("#cdd6f4")
	ColorSubtext  = lipgloss.Color("#a6adc8")
	ColorLavender = lipgloss.Color("#b4befe")
	ColorBlue     = lipgloss.Color("#89b4fa")
	ColorSapphire = lipgloss.Color("#74c7ec")
	ColorGreen    = lipgloss.Color("#a6e3a1")
	ColorYellow   = lipgloss.Color("#f9e2af")
	ColorPeach    = lipgloss.Color("#fab387")
	ColorRed      = lipgloss.Color("#f38ba8")
	ColorMauve    = lipgloss.Color("#cba6f7")
	ColorPink     = lipgloss.Color("#f5c2e7")
)

// Chart bar colors cycling.
var BarColors = []lipgloss.Color{
	ColorBlue, ColorGreen, ColorMauve, ColorPeach, ColorSapphire, ColorYellow,
}

var (
	StyleApp = lipgloss.NewStyle().
			Background(ColorBase)

	StyleTitle = lipgloss.NewStyle().
			Foreground(ColorLavender).
			Bold(true).
			Padding(0, 1)

	StyleSubtitle = lipgloss.NewStyle().
			Foreground(ColorSubtext).
			Italic(true)

	StyleStatBox = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(ColorOverlay).
			Padding(0, 1).
			MarginRight(1)

	StyleStatLabel = lipgloss.NewStyle().
			Foreground(ColorSubtext).
			Bold(true)

	StyleStatValue = lipgloss.NewStyle().
			Foreground(ColorText).
			Bold(true)

	StyleStatCost = lipgloss.NewStyle().
			Foreground(ColorGreen).
			Bold(true)

	StyleTab = lipgloss.NewStyle().
			Foreground(ColorSubtext).
			Padding(0, 2)

	StyleActiveTab = lipgloss.NewStyle().
			Foreground(ColorLavender).
			Bold(true).
			Padding(0, 2).
			Underline(true)

	StyleHelp = lipgloss.NewStyle().
			Foreground(ColorOverlay)

	StyleTableHeader = lipgloss.NewStyle().
			Foreground(ColorLavender).
			Bold(true)

	StyleTableRow = lipgloss.NewStyle().
			Foreground(ColorText)

	StyleTableRowAlt = lipgloss.NewStyle().
			Foreground(ColorSubtext)

	StyleSelectedRow = lipgloss.NewStyle().
			Foreground(ColorBase).
			Background(ColorLavender).
			Bold(true)

	StyleSectionTitle = lipgloss.NewStyle().
			Foreground(ColorBlue).
			Bold(true).
			MarginTop(1)

	StyleWarning = lipgloss.NewStyle().
			Foreground(ColorYellow)

	StyleError = lipgloss.NewStyle().
			Foreground(ColorRed)

	StylePositive = lipgloss.NewStyle().
			Foreground(ColorGreen)

	StyleNegative = lipgloss.NewStyle().
			Foreground(ColorRed)

	StyleMuted = lipgloss.NewStyle().
			Foreground(ColorOverlay)
)
