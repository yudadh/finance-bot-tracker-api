package service

import (
	"context"
	"testing"
	"time"

	"github.com/yudadh/finance-bot-tracker-api/internal/domain"
)

type fakeBudgetRepo struct {
	budget *domain.Budget
	err error
}

type fakeTransactionSummarizer struct {
	totalTransactionAmount int64
	err error
}

func (f *fakeBudgetRepo) Create(ctx context.Context, budget *domain.Budget) error {
	return f.err
}

func (f *fakeBudgetRepo) FindByUserPeriod(
	ctx context.Context,
	userID uint64,
	period time.Time,
	periodType domain.BudgetPeriodType,
) (*domain.Budget, error) {
	return f.budget, f.err
}

func (f *fakeBudgetRepo) Update(ctx context.Context, budget *domain.Budget) error {
	f.budget = budget
	return f.err
}

func (f *fakeBudgetRepo) FindByID(ctx context.Context, budgetID uint64) (*domain.Budget, error) {
	return f.budget, f.err
}


func (f *fakeTransactionSummarizer) SumTransactionsByUserAndDateRange(
	ctx context.Context, 
	userID uint64,
	transactionType domain.TransactionType, 
	startDate time.Time, 
	endDate time.Time,
) (*int64, error) {
	return &f.totalTransactionAmount, f.err
}

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

func TestCheckBudgetAlert_AlertFires(t *testing.T) {
	tests := []struct{
		name string
		amount uint64
		totalTransactionAmount int64
		shouldAlert bool
		threshold uint8
		usedPercentage uint64
	}{
		{
			name: "80 alert fires",
			amount: 100000,
			totalTransactionAmount: 90000,
			shouldAlert: true,
			threshold: 80,
			usedPercentage: 90,
		},
		{
			name: "100 alert fires",
			amount: 100000,
			totalTransactionAmount: 100000,
			shouldAlert: true,
			threshold: 100,
			usedPercentage: 100,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ctx := context.Background()
		
			loc, _ := time.LoadLocation("Asia/Makassar")
			startDate := time.Date(2026, 8, 28, 0, 0, 0, 0, loc)
			budgetRepo := &fakeBudgetRepo{budget: &domain.Budget{
				ID: 1,
				UserID: 1,
				PeriodType: domain.BudgetPeriodTypeMonthly,
				PeriodStart: startDate,
				Amount: test.amount,
				Currency: "IDR",
			}}
			transactionRepo := &fakeTransactionSummarizer{totalTransactionAmount: test.totalTransactionAmount}
		
			service := NewBudgetService(budgetRepo, transactionRepo, nil)
		
			res, _ := service.CheckBudgetAlert(ctx, BudgetStatusInput{
				UserID: 1,
				PeriodType: domain.BudgetPeriodTypeMonthly,
				Location: loc,
			})
		
			if res.ShouldAlert != test.shouldAlert {
				t.Fatalf("expected %t, got %t", test.shouldAlert, res.ShouldAlert)
			}
		
			if res.Threshold != test.threshold {
				t.Fatalf("expected %d, got %d", test.threshold, res.Threshold)
			}
		
			if res.UsedPercentage != test.usedPercentage {
				t.Fatalf("expected %d, got %d", test.usedPercentage, res.UsedPercentage)
			}
		})
	}

}

func TestCheckBudgetAlert_DoesNotFireWhenAlreadySent(t *testing.T) {
	tests := []struct{
		name string
		amount uint64
		totalTransactionAmount int64
	}{
		{
			name: "should not alert fires at 80",
			amount: 100000,
			totalTransactionAmount: 90000,
		},
		{
			name: "should not alert fires at 100",
			amount: 100000,
			totalTransactionAmount: 100000,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ctx := context.Background()
		
			loc, _ := time.LoadLocation("Asia/Makassar")
			startDate := time.Date(2026, 8, 28, 0, 0, 0, 0, loc)
			alert80SentAt := time.Date(2026, 8, 28, 15, 0, 0, 0, loc)
			alert100SentAt := time.Date(2026, 8, 28, 16, 0, 0, 0, loc)
			budgetRepo := &fakeBudgetRepo{budget: &domain.Budget{
				ID: 1,
				UserID: 1,
				PeriodType: domain.BudgetPeriodTypeMonthly,
				PeriodStart: startDate,
				Amount: test.amount,
				Currency: "IDR",
				Alert80SentAt: &alert80SentAt,
				Alert100SentAt: &alert100SentAt,
			}}
			transactionRepo := &fakeTransactionSummarizer{totalTransactionAmount: test.totalTransactionAmount}
		
			service := NewBudgetService(budgetRepo, transactionRepo, nil)
		
			res, _ := service.CheckBudgetAlert(ctx, BudgetStatusInput{
				UserID: 1,
				PeriodType: domain.BudgetPeriodTypeMonthly,
				Location: loc,
			})
		
			if res.ShouldAlert {
				t.Fatalf("expected false, got %t", res.ShouldAlert)
			}

		})
	}
}

func TestCheckBudgetAlert_DoesNotFireWhenBudget0(t *testing.T) {
	ctx := context.Background()
		
	loc, _ := time.LoadLocation("Asia/Makassar")
	startDate := time.Date(2026, 8, 28, 0, 0, 0, 0, loc)
	budgetRepo := &fakeBudgetRepo{budget: &domain.Budget{
		ID: 1,
		UserID: 1,
		PeriodType: domain.BudgetPeriodTypeMonthly,
		PeriodStart: startDate,
		Amount: 0,
		Currency: "IDR",
	}}
	transactionRepo := &fakeTransactionSummarizer{totalTransactionAmount: 10000}

	service := NewBudgetService(budgetRepo, transactionRepo, nil)

	res, _ := service.CheckBudgetAlert(ctx, BudgetStatusInput{
		UserID: 1,
		PeriodType: domain.BudgetPeriodTypeMonthly,
		Location: loc,
	})

	if res.ShouldAlert {
		t.Fatalf("expected false, got %t", res.ShouldAlert)
	}
}