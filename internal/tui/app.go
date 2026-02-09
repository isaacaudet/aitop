package tui

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/isaacaudet/clawdtop/internal/config"
	"github.com/isaacaudet/clawdtop/internal/model"
	"github.com/isaacaudet/clawdtop/internal/provider"
)

type viewType int

const (
	viewDashboard viewType = iota
	viewSessions
	viewProviders
	viewHeatmap
	viewLive
)

var viewNames = []string{"Dashboard", "Sessions", "Providers", "Heatmap", "Live"}

// sortMode defines session sort modes.
type sortMode int

const (
	sortByDate sortMode = iota
	sortByCost
	sortByTokens
	sortByDuration
)

var sortModeNames = []string{"date", "cost", "tokens", "duration"}

// timePeriod defines time period filter.
type timePeriod int

const (
	periodAllTime timePeriod = iota
	periodThisMonth
	periodThisWeek
	periodToday
)

var timePeriodNames = []string{"All Time", "This Month", "This Week", "Today"}

// Model is the main Bubble Tea model.
type Model struct {
	cache     *model.StatsCache
	aggData   *provider.AggregatedData
	providers []provider.Provider
	cfg       config.Config
	view      viewType
	width     int
	height    int
	showHelp  bool

	sessView   sessionsView
	viewport   viewport.Model
	sortMode   sortMode
	timePeriod timePeriod

	// Live view state
	liveView   liveView
}

type dataLoadedMsg struct {
	aggData *provider.AggregatedData
}

type tickMsg time.Time

// New creates a new TUI model.
func New(cache *model.StatsCache, providers []provider.Provider) Model {
	cfg := config.Load()
	vp := viewport.New(80, 40)
	return Model{
		cache:     cache,
		providers: providers,
		cfg:       cfg,
		view:      viewDashboard,
		sessView:  newSessionsView(nil),
		viewport:  vp,
		liveView:  newLiveView(),
	}
}

func loadDataCmd(providers []provider.Provider) tea.Cmd {
	return func() tea.Msg {
		agg := provider.LoadAll(providers)
		return dataLoadedMsg{aggData: agg}
	}
}

func tickCmd() tea.Cmd {
	return tea.Tick(5*time.Second, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(loadDataCmd(m.providers), tickCmd())
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		headerHeight := 5 // title + cost + tabs + padding
		footerHeight := 2
		m.viewport.Width = msg.Width
		m.viewport.Height = msg.Height - headerHeight - footerHeight
		if m.viewport.Height < 10 {
			m.viewport.Height = 10
		}
		return m, nil

	case dataLoadedMsg:
		m.aggData = msg.aggData
		// Build session view from all provider sessions.
		var sessions []provider.SessionInfo
		if m.aggData != nil {
			for _, p := range m.aggData.Providers {
				sessions = append(sessions, p.Sessions...)
			}
		}
		m.sessView = newSessionsViewFromProvider(sessions)
		m.sessView.sort = m.sortMode
		m.sessView.applySort()
		return m, nil

	case tickMsg:
		if m.view == viewLive {
			return m, tea.Batch(loadDataCmd(m.providers), tickCmd())
		}
		return m, tickCmd()

	case tea.KeyMsg:
		switch {
		case key.Matches(msg, keys.Quit):
			return m, tea.Quit
		case key.Matches(msg, keys.Help):
			m.showHelp = !m.showHelp
			return m, nil
		case key.Matches(msg, keys.Tab):
			m.view = (m.view + 1) % viewType(len(viewNames))
			m.viewport.GotoTop()
			return m, nil
		case key.Matches(msg, keys.View1):
			m.view = viewDashboard
			m.viewport.GotoTop()
		case key.Matches(msg, keys.View2):
			m.view = viewSessions
			m.viewport.GotoTop()
		case key.Matches(msg, keys.View3):
			m.view = viewProviders
			m.viewport.GotoTop()
		case key.Matches(msg, keys.View4):
			m.view = viewHeatmap
			m.viewport.GotoTop()
		case key.Matches(msg, keys.View5):
			m.view = viewLive
			m.viewport.GotoTop()
		case key.Matches(msg, keys.Escape):
			if m.showHelp {
				m.showHelp = false
				return m, nil
			}
			if m.view == viewSessions && m.sessView.expanded {
				m.sessView.expanded = false
				return m, nil
			}
		case key.Matches(msg, keys.Refresh):
			return m, loadDataCmd(m.providers)
		case key.Matches(msg, keys.Sort):
			if m.view == viewSessions {
				m.sortMode = (m.sortMode + 1) % sortMode(len(sortModeNames))
				m.sessView.sort = m.sortMode
				m.sessView.applySort()
			}
			return m, nil
		case key.Matches(msg, keys.TimePeriod):
			if m.view == viewProviders {
				m.timePeriod = (m.timePeriod + 1) % timePeriod(len(timePeriodNames))
			}
			return m, nil
		}

		// Viewport scrolling for non-sessions views.
		if m.view != viewSessions {
			switch {
			case key.Matches(msg, keys.Down):
				m.viewport.LineDown(1)
				return m, nil
			case key.Matches(msg, keys.Up):
				m.viewport.LineUp(1)
				return m, nil
			case key.Matches(msg, keys.HalfPgDown):
				m.viewport.HalfViewDown()
				return m, nil
			case key.Matches(msg, keys.HalfPgUp):
				m.viewport.HalfViewUp()
				return m, nil
			}
		}

		// Live view h/l scroll.
		if m.view == viewLive {
			switch {
			case key.Matches(msg, keys.Left):
				m.liveView.scrollLeft()
				return m, nil
			case key.Matches(msg, keys.Right):
				m.liveView.scrollRight()
				return m, nil
			}
		}

		// Session view keys.
		if m.view == viewSessions {
			switch {
			case key.Matches(msg, keys.Down):
				if m.sessView.selected < len(m.sessView.sessions)-1 {
					m.sessView.selected++
					if m.sessView.selected >= m.sessView.scroll+m.sessView.maxVisible {
						m.sessView.scroll++
					}
				}
			case key.Matches(msg, keys.Up):
				if m.sessView.selected > 0 {
					m.sessView.selected--
					if m.sessView.selected < m.sessView.scroll {
						m.sessView.scroll--
					}
				}
			case key.Matches(msg, keys.HalfPgDown):
				jump := m.sessView.maxVisible / 2
				m.sessView.selected += jump
				if m.sessView.selected >= len(m.sessView.sessions) {
					m.sessView.selected = len(m.sessView.sessions) - 1
				}
				if m.sessView.selected >= m.sessView.scroll+m.sessView.maxVisible {
					m.sessView.scroll = m.sessView.selected - m.sessView.maxVisible + 1
				}
			case key.Matches(msg, keys.HalfPgUp):
				jump := m.sessView.maxVisible / 2
				m.sessView.selected -= jump
				if m.sessView.selected < 0 {
					m.sessView.selected = 0
				}
				if m.sessView.selected < m.sessView.scroll {
					m.sessView.scroll = m.sessView.selected
				}
			case key.Matches(msg, keys.Enter):
				m.sessView.expanded = true
			}
		}
	}

	return m, nil
}

func (m Model) View() string {
	if m.width == 0 {
		return "Loading..."
	}

	var sb strings.Builder

	// Header with provider icons.
	title := StyleTitle.Render("  clawdtop")
	subtitle := StyleSubtitle.Render(" — AI Usage Dashboard")
	providerIcons := ""
	if m.aggData != nil {
		for _, p := range m.aggData.Providers {
			iconStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(p.Color))
			providerIcons += " " + iconStyle.Render(p.Icon+" "+p.ProviderName)
		}
	}
	sb.WriteString(lipgloss.JoinHorizontal(lipgloss.Center, title, subtitle, "  ", providerIcons))
	sb.WriteString("\n")

	// Plan usage banner.
	if m.cfg.Plan.MonthlyCost > 0 && m.aggData != nil {
		now := time.Now()
		monthStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location()).Format("2006-01-02")
		var monthUsage float64
		for _, d := range m.aggData.DailyUsage {
			if d.Date >= monthStart {
				monthUsage += d.Cost
			}
		}
		pct := monthUsage / m.cfg.Plan.MonthlyCost * 100
		planStr := fmt.Sprintf("  %s $%.0f/mo — $%.2f (%.0f%%)",
			m.cfg.Plan.Name, m.cfg.Plan.MonthlyCost, monthUsage, pct)

		var planStyle lipgloss.Style
		switch {
		case pct >= 90:
			planStyle = lipgloss.NewStyle().Foreground(ColorRed).Bold(true)
		case pct >= 70:
			planStyle = lipgloss.NewStyle().Foreground(ColorYellow).Bold(true)
		default:
			planStyle = lipgloss.NewStyle().Foreground(ColorGreen).Bold(true)
		}
		sb.WriteString(planStyle.Render(planStr))
		sb.WriteString("\n")
	} else if m.aggData != nil && m.aggData.TotalCost > 0 {
		// Fallback: total spend banner.
		costStr := fmt.Sprintf("$%.2f", m.aggData.TotalCost)
		banner := lipgloss.NewStyle().
			Foreground(ColorGreen).
			Bold(true).
			Render(fmt.Sprintf("  Total AI Spend: %s", costStr))
		sb.WriteString(banner)
		sb.WriteString("\n")
	}

	// Tab bar.
	var tabs []string
	for i, name := range viewNames {
		num := fmt.Sprintf("%d", i+1)
		label := fmt.Sprintf(" %s %s ", num, name)
		if viewType(i) == m.view {
			tabs = append(tabs, StyleActiveTab.Render(label))
		} else {
			tabs = append(tabs, StyleTab.Render(label))
		}
	}
	// Add sort/filter indicators.
	if m.view == viewSessions {
		tabs = append(tabs, StyleMuted.Render(fmt.Sprintf(" [sort: %s]", sortModeNames[m.sortMode])))
	}
	if m.view == viewProviders {
		tabs = append(tabs, StyleMuted.Render(fmt.Sprintf(" [%s]", timePeriodNames[m.timePeriod])))
	}
	sb.WriteString(lipgloss.JoinHorizontal(lipgloss.Top, tabs...))
	sb.WriteString("\n\n")

	if m.showHelp {
		sb.WriteString(helpView())
		sb.WriteString("\n\n")
	}

	contentWidth := m.width - 2
	if contentWidth < 40 {
		contentWidth = 40
	}

	// Render content through viewport for scrollable views.
	var content string
	switch m.view {
	case viewDashboard:
		content = renderDashboard(m.cache, m.aggData, contentWidth, m.cfg)
	case viewSessions:
		content = m.sessView.render(contentWidth)
	case viewProviders:
		content = renderProvidersView(m.aggData, contentWidth, m.timePeriod)
	case viewHeatmap:
		content = renderHeatmap(m.aggData, contentWidth)
	case viewLive:
		content = m.liveView.render(m.aggData, contentWidth)
	}

	if m.view == viewSessions {
		// Sessions has its own scroll handling.
		// Cap visible rows to fit terminal.
		availableHeight := m.height - 8 // header + tabs + footer
		if availableHeight > 5 && m.sessView.maxVisible > availableHeight-3 {
			m.sessView.maxVisible = availableHeight - 3
		}
		sb.WriteString(content)
	} else {
		// Use viewport for other views.
		m.viewport.SetContent(content)
		sb.WriteString(m.viewport.View())
	}

	sb.WriteString("\n")
	footer := StyleHelp.Render(" tab: views | 1-5: jump | j/k: scroll | ctrl+u/d: half-page | s: sort | t: period | ?: help | q: quit")
	sb.WriteString(footer)

	return sb.String()
}
