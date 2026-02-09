package model

import "time"

// AggregateDaily converts StatsCache data into a slice of DailyStats with costs.
func AggregateDaily(cache *StatsCache) []DailyStats {
	// Index token data by date.
	tokensByDate := make(map[string]map[string]int)
	for _, dt := range cache.DailyModelTokens {
		tokensByDate[dt.Date] = dt.TokensByModel
	}

	var days []DailyStats
	for _, da := range cache.DailyActivity {
		ds := DailyStats{
			Date:          da.Date,
			Messages:      da.MessageCount,
			Sessions:      da.SessionCount,
			ToolCalls:     da.ToolCallCount,
			TokensByModel: tokensByDate[da.Date],
		}
		var totalTokens int
		var cost float64
		if models, ok := tokensByDate[da.Date]; ok {
			for m, tokens := range models {
				totalTokens += tokens
				// Estimate cost: dailyModelTokens only gives total tokens, not split.
				// Use output pricing as an approximation since these are output tokens.
				if p, ok := GetPricing(m); ok {
					cost += float64(tokens) * p.OutputPerMTok / 1_000_000
				}
			}
		}
		ds.TotalTokens = totalTokens
		ds.Cost = cost
		days = append(days, ds)
	}
	return days
}

// PeriodFromDays aggregates DailyStats into a PeriodSummary.
func PeriodFromDays(label string, days []DailyStats) PeriodSummary {
	ps := PeriodSummary{
		Label:      label,
		Days:       len(days),
		ModelCosts: make(map[string]float64),
	}
	for _, d := range days {
		ps.Messages += d.Messages
		ps.Sessions += d.Sessions
		ps.ToolCalls += d.ToolCalls
		ps.TotalTokens += d.TotalTokens
		ps.Cost += d.Cost
	}
	return ps
}

// FilterDaysAfter returns daily stats on or after the given date string (YYYY-MM-DD).
func FilterDaysAfter(days []DailyStats, after string) []DailyStats {
	var result []DailyStats
	for _, d := range days {
		if d.Date >= after {
			result = append(result, d)
		}
	}
	return result
}

// ComputeSummaries returns Today, This Week, This Month, and All Time summaries.
func ComputeSummaries(cache *StatsCache) (today, week, month, allTime PeriodSummary) {
	days := AggregateDaily(cache)
	now := time.Now()
	todayStr := now.Format("2006-01-02")
	weekAgo := now.AddDate(0, 0, -7).Format("2006-01-02")
	monthAgo := now.AddDate(0, -1, 0).Format("2006-01-02")

	today = PeriodFromDays("Today", FilterDaysAfter(days, todayStr))
	week = PeriodFromDays("This Week", FilterDaysAfter(days, weekAgo))
	month = PeriodFromDays("This Month", FilterDaysAfter(days, monthAgo))
	allTime = PeriodFromDays("All Time", days)

	// Override all-time cost with precise calculation from modelUsage.
	allTime.Cost = TotalCostFromModelUsage(cache.ModelUsage)

	return
}

// ComputeBurnRate calculates spending rate from daily stats.
func ComputeBurnRate(days []DailyStats) BurnRate {
	if len(days) == 0 {
		return BurnRate{}
	}

	now := time.Now()
	weekAgo := now.AddDate(0, 0, -7).Format("2006-01-02")
	twoWeeksAgo := now.AddDate(0, 0, -14).Format("2006-01-02")

	thisWeek := FilterDaysAfter(days, weekAgo)
	var lastWeek []DailyStats
	for _, d := range days {
		if d.Date >= twoWeeksAgo && d.Date < weekAgo {
			lastWeek = append(lastWeek, d)
		}
	}

	var thisWeekCost, lastWeekCost float64
	for _, d := range thisWeek {
		thisWeekCost += d.Cost
	}
	for _, d := range lastWeek {
		lastWeekCost += d.Cost
	}

	activeDays := len(thisWeek)
	if activeDays == 0 {
		activeDays = 1
	}

	dailyAvg := thisWeekCost / float64(activeDays)
	projected := dailyAvg * 30

	var trend float64
	if lastWeekCost > 0 {
		trend = ((thisWeekCost - lastWeekCost) / lastWeekCost) * 100
	}

	return BurnRate{
		DailyAvg:        dailyAvg,
		ProjectedMonth:  projected,
		TrendVsLastWeek: trend,
	}
}
