package parser

import (
	"strings"
	"time"
)

func parseSimpleDate(text string, now time.Time, location *time.Location) time.Time {
	if location == nil {
		location = time.Local
	}

	current := now.In(location)

	if strings.Contains(text, "yesterday") || strings.Contains(text, "kemarin") {
		return dateOnly(current.AddDate(0, 0, -1))
	}

	return dateOnly(current)
}

func dateOnly(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
}