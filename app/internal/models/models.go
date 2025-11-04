package models

import "time"

type URL struct {
	ID        int64     `json:"id"`
	Original  string    `json:"original"`
	Shortened string    `json:"shortened"`
	CreatedAt time.Time `json:"created_at"`
}

type Click struct {
	ID        int64     `json:"id"`
	URLID     int64     `json:"url_id"`
	ClickedAt time.Time `json:"clicked_at"`
	UserAgent string    `json:"user_agent"`
}

type By string

const (
	ByDay       By = "day"
	ByMonth     By = "month"
	ByYear      By = "year"
	ByUserAgent By = "user_agent"
)

type AnalyticsParams struct {
	URL     string    `json:"url"`
	From    time.Time `json:"from"`
	To      time.Time `json:"to"`
	GroupBy By        `json:"group_by"`
}

type AnalyticsResult struct {
	TotalClicks int            `json:"total_clicks"`
	Grouped     map[string]int `json:"grouped"`
}
