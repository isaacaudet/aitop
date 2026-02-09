package tui

import "github.com/charmbracelet/bubbles/key"

type keyMap struct {
	Quit       key.Binding
	Tab        key.Binding
	View1      key.Binding
	View2      key.Binding
	View3      key.Binding
	View4      key.Binding
	View5      key.Binding
	Up         key.Binding
	Down       key.Binding
	Enter      key.Binding
	Escape     key.Binding
	Refresh    key.Binding
	Help       key.Binding
	Sort       key.Binding
	TimePeriod key.Binding
	HalfPgUp   key.Binding
	HalfPgDown key.Binding
	Left       key.Binding
	Right      key.Binding
}

var keys = keyMap{
	Quit: key.NewBinding(
		key.WithKeys("q", "ctrl+c"),
		key.WithHelp("q", "quit"),
	),
	Tab: key.NewBinding(
		key.WithKeys("tab"),
		key.WithHelp("tab", "next view"),
	),
	View1: key.NewBinding(
		key.WithKeys("1"),
		key.WithHelp("1", "dashboard"),
	),
	View2: key.NewBinding(
		key.WithKeys("2"),
		key.WithHelp("2", "sessions"),
	),
	View3: key.NewBinding(
		key.WithKeys("3"),
		key.WithHelp("3", "providers"),
	),
	View4: key.NewBinding(
		key.WithKeys("4"),
		key.WithHelp("4", "heatmap"),
	),
	View5: key.NewBinding(
		key.WithKeys("5"),
		key.WithHelp("5", "live"),
	),
	Up: key.NewBinding(
		key.WithKeys("k", "up"),
		key.WithHelp("k/↑", "up"),
	),
	Down: key.NewBinding(
		key.WithKeys("j", "down"),
		key.WithHelp("j/↓", "down"),
	),
	Enter: key.NewBinding(
		key.WithKeys("enter"),
		key.WithHelp("enter", "expand"),
	),
	Escape: key.NewBinding(
		key.WithKeys("esc"),
		key.WithHelp("esc", "collapse"),
	),
	Refresh: key.NewBinding(
		key.WithKeys("r"),
		key.WithHelp("r", "refresh"),
	),
	Help: key.NewBinding(
		key.WithKeys("?"),
		key.WithHelp("?", "help"),
	),
	Sort: key.NewBinding(
		key.WithKeys("s"),
		key.WithHelp("s", "sort mode"),
	),
	TimePeriod: key.NewBinding(
		key.WithKeys("t"),
		key.WithHelp("t", "time period"),
	),
	HalfPgUp: key.NewBinding(
		key.WithKeys("ctrl+u"),
		key.WithHelp("ctrl+u", "half page up"),
	),
	HalfPgDown: key.NewBinding(
		key.WithKeys("ctrl+d"),
		key.WithHelp("ctrl+d", "half page down"),
	),
	Left: key.NewBinding(
		key.WithKeys("h", "left"),
		key.WithHelp("h/←", "left"),
	),
	Right: key.NewBinding(
		key.WithKeys("l", "right"),
		key.WithHelp("l/→", "right"),
	),
}

func helpView() string {
	help := StyleSectionTitle.Render("Keyboard Shortcuts") + "\n\n"
	bindings := []struct{ key, desc string }{
		{"tab", "Cycle views"},
		{"1-5", "Jump to view"},
		{"j/k", "Scroll up/down"},
		{"ctrl+u/d", "Half page up/down"},
		{"enter", "Expand selection"},
		{"esc", "Collapse / close help"},
		{"s", "Cycle sort mode (sessions)"},
		{"t", "Cycle time period (providers)"},
		{"h/l", "Scroll left/right (live)"},
		{"r", "Refresh data"},
		{"?", "Toggle help"},
		{"q", "Quit"},
	}
	for _, b := range bindings {
		help += StyleActiveTab.Render(b.key) + "  " + StyleStatLabel.Render(b.desc) + "\n"
	}
	return help
}
