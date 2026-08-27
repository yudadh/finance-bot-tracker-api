package service

import (
	"testing"
	"time"

	"github.com/yudadh/finance-bot-tracker-api/internal/domain"
)

func TestSetBudgetWindowWeeklyUsesLocalWeekday(t *testing.T) {
	loc := time.FixedZone("WITA", 8*60*60)
	now := time.Date(2026, 8, 23, 16, 30, 0, 0, time.UTC)

	startDate, endDate := setBudgetWindow(now, loc, domain.BudgetPeriodTypeWeekly)

	if startDate.Format("2006-01-02 15:04 MST") != "2026-08-24 00:00 WITA" {
		t.Fatalf("expected weekly start in local Monday, got %s", startDate.Format("2006-01-02 15:04 MST"))
	}

	if endDate.Format("2006-01-02 15:04 MST") != "2026-08-31 00:00 WITA" {
		t.Fatalf("expected weekly end on next local Monday, got %s", endDate.Format("2006-01-02 15:04 MST"))
	}
}
