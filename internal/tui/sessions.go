package tui

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/isaacaudet/aitop/internal/provider"
	"github.com/isaacaudet/aitop/internal/tui/components"
)

type sessionsView struct {
	sessions   []provider.SessionInfo
	selected   int
	scroll     int
	expanded   bool
	maxVisible int
	sort       sortMode
}

func newSessionsView(sessions []provider.SessionInfo) sessionsView {
	return newSessionsViewFromProvider(sessions)
}

func newSessionsViewFromProvider(sessions []provider.SessionInfo) sessionsView {
	sv := sessionsView{
		sessions:   sessions,
		maxVisible: 20,
		sort:       sortByDate,
	}
	sv.applySort()
	return sv
}

func (sv *sessionsView) applySort() {
	switch sv.sort {
	case sortByDate:
		sort.Slice(sv.sessions, func(i, j int) bool {
			return sv.sessions[i].StartTime.After(sv.sessions[j].StartTime)
		})
	case sortByCost:
		sort.Slice(sv.sessions, func(i, j int) bool {
			return sv.sessions[i].Cost > sv.sessions[j].Cost
		})
	case sortByTokens:
		sort.Slice(sv.sessions, func(i, j int) bool {
			return sv.sessions[i].Tokens > sv.sessions[j].Tokens
		})
	case sortByDuration:
		sort.Slice(sv.sessions, func(i, j int) bool {
			di := sv.sessions[i].EndTime.Sub(sv.sessions[i].StartTime)
			dj := sv.sessions[j].EndTime.Sub(sv.sessions[j].StartTime)
			return di > dj
		})
	}
}

func (sv sessionsView) render(width int) string {
	if len(sv.sessions) == 0 {
		return StyleMuted.Render("  No session data loaded.")
	}

	var sb strings.Builder

	if !sv.expanded {
		// Group sessions by project.
		type projectGroup struct {
			name     string
			sessions []int // indices into sv.sessions
			cost     float64
		}

		projectOrder := []string{}
		projectMap := make(map[string]*projectGroup)
		for i, s := range sv.sessions {
			proj := s.Project
			if proj == "" {
				proj = "(unknown)"
			}
			if _, ok := projectMap[proj]; !ok {
				projectMap[proj] = &projectGroup{name: proj}
				projectOrder = append(projectOrder, proj)
			}
			pg := projectMap[proj]
			pg.sessions = append(pg.sessions, i)
			pg.cost += s.Cost
		}

		columns := []components.Column{
			{Title: "Project", Width: 22},
			{Title: "Date", Width: 14},
			{Title: "Duration", Width: 10},
			{Title: "Msgs", Width: 8, Align: 1},
			{Title: "Tokens", Width: 12, Align: 1},
			{Title: "Cost", Width: 10, Align: 1},
			{Title: "Model", Width: 16},
		}

		var rows [][]string
		for _, s := range sv.sessions {
			duration := s.EndTime.Sub(s.StartTime)
			msgStr := fmt.Sprintf("%d", s.Messages)
			if s.UserMessages > 0 {
				msgStr = fmt.Sprintf("%d/%d", s.UserMessages, s.Messages)
			}
			rows = append(rows, []string{
				truncate(s.Project, 22),
				s.StartTime.Format("Jan 02 15:04"),
				formatDuration(duration),
				msgStr,
				components.FormatTokens(s.Tokens),
				fmt.Sprintf("$%.2f", s.Cost),
				truncate(s.Model, 16),
			})
		}

		table := components.Table{
			Columns:      columns,
			Rows:         rows,
			Selected:     sv.selected,
			ScrollOffset: sv.scroll,
			MaxVisible:   sv.maxVisible,
		}
		sb.WriteString(table.Render())

		// Project summary below (only show top 5 to avoid overflowing terminal).
		if len(projectOrder) > 1 {
			sb.WriteString("\n")
			sb.WriteString(StyleSectionTitle.Render("Projects"))
			sb.WriteString("\n")
			// Sort projects by cost.
			sort.Slice(projectOrder, func(i, j int) bool {
				return projectMap[projectOrder[i]].cost > projectMap[projectOrder[j]].cost
			})
			maxLabelLen := 0
			for _, name := range projectOrder {
				if len(name) > maxLabelLen {
					maxLabelLen = len(name)
				}
			}
			if maxLabelLen > 30 {
				maxLabelLen = 30
			}
			var maxProjCost float64
			for _, pg := range projectMap {
				if pg.cost > maxProjCost {
					maxProjCost = pg.cost
				}
			}
			barWidth := width - maxLabelLen - 20
			if barWidth < 10 {
				barWidth = 10
			}
			shown := projectOrder
			if len(shown) > 5 {
				shown = shown[:5]
			}
			for i, name := range shown {
				pg := projectMap[name]
				if pg.cost < 0.01 {
					continue
				}
				color := BarColors[i%len(BarColors)]
				displayName := truncate(name, maxLabelLen)
				bar := HorizontalBarAligned(displayName, pg.cost, maxProjCost, barWidth, maxLabelLen, color)
				sb.WriteString(bar)
				sb.WriteString(StyleStatCost.Render(fmt.Sprintf("  $%.2f", pg.cost)))
				sb.WriteString(StyleMuted.Render(fmt.Sprintf("  %d sess", len(pg.sessions))))
				sb.WriteString("\n")
			}
		}
	} else {
		if sv.selected >= 0 && sv.selected < len(sv.sessions) {
			s := sv.sessions[sv.selected]
			sb.WriteString(renderSessionDetailFromProvider(s))
		}
	}

	return sb.String()
}

func renderSessionDetailFromProvider(s provider.SessionInfo) string {
	var sb strings.Builder

	sb.WriteString(StyleSectionTitle.Render("Session Detail"))
	sb.WriteString("\n\n")
	sb.WriteString(fmt.Sprintf("  ID:       %s\n", StyleStatValue.Render(s.ID)))
	sb.WriteString(fmt.Sprintf("  Project:  %s\n", StyleStatValue.Render(s.Project)))
	if s.Model != "" {
		sb.WriteString(fmt.Sprintf("  Model:    %s\n", StyleStatValue.Render(s.Model)))
	}
	sb.WriteString(fmt.Sprintf("  Start:    %s\n", StyleStatValue.Render(s.StartTime.Format(time.RFC3339))))
	duration := s.EndTime.Sub(s.StartTime)
	if duration > 0 {
		sb.WriteString(fmt.Sprintf("  Duration: %s\n", StyleStatValue.Render(formatDuration(duration))))
	}
	if s.Messages > 0 {
		msgStr := fmt.Sprintf("%d", s.Messages)
		if s.UserMessages > 0 {
			msgStr = fmt.Sprintf("%d total (%d user, %d assistant)", s.Messages, s.UserMessages, s.Messages-s.UserMessages)
		}
		sb.WriteString(fmt.Sprintf("  Messages: %s\n", StyleStatValue.Render(msgStr)))
	}
	if s.Tokens > 0 {
		sb.WriteString(fmt.Sprintf("  Tokens:   %s\n", StyleStatValue.Render(components.FormatTokens(s.Tokens))))
	}
	if s.Cost > 0 {
		sb.WriteString(fmt.Sprintf("  Cost:     %s\n", StyleStatCost.Render(fmt.Sprintf("$%.2f", s.Cost))))
	}

	sb.WriteString("\n")
	sb.WriteString(StyleMuted.Render("  Press esc to go back"))
	sb.WriteString("\n")

	return sb.String()
}

func formatDuration(d time.Duration) string {
	if d <= 0 {
		return "--"
	}
	if d < time.Minute {
		return fmt.Sprintf("%ds", int(d.Seconds()))
	}
	if d < time.Hour {
		return fmt.Sprintf("%dm%ds", int(d.Minutes()), int(d.Seconds())%60)
	}
	return fmt.Sprintf("%dh%dm", int(d.Hours()), int(d.Minutes())%60)
}

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max-1] + "…"
}
