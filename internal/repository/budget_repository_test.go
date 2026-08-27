package repository

import (
	"testing"
	"time"
)

func TestBudgetPeriodDateKeepsLocalCalendarDate(t *testing.T) {
	loc := time.FixedZone("WITA", 8*60*60)
	period := time.Date(2026, 8, 24, 0, 0, 0, 0, loc)

	result := budgetPeriodDate(period)

	if result.Location() != time.UTC {
		t.Fatalf("expected UTC location, got %v", result.Location())
	}

	if result.Format("2006-01-02") != "2026-08-24" {
		t.Fatalf("expected date 2026-08-24, got %s", result.Format("2006-01-02"))
	}
}
