package repository

import (
	"context"
	"time"

	"github.com/yudadh/finance-bot-tracker-api/internal/domain"
	"gorm.io/gorm"
)

type TransactionRepository struct {
	db *gorm.DB
}

func NewTransactionRepository(db *gorm.DB) *TransactionRepository {
	return &TransactionRepository{db: db}
}

func (r *TransactionRepository) Create(ctx context.Context, transaction *domain.Transaction) error {
	return r.db.WithContext(ctx).Create(transaction).Error
}

func (r *TransactionRepository) Update(ctx context.Context, transaction *domain.Transaction) error {
	return r.db.WithContext(ctx).Save(transaction).Error
}

func (r *TransactionRepository) FindByUserAndDateRange(
	ctx context.Context, 
	userID uint64, 
	startDate time.Time, 
	endDate time.Time,
	limit int,
	offset int,
) ([]domain.Transaction, error) {
	var transactions []domain.Transaction

	err := r.db.
		WithContext(ctx).
		Where("user_id = ?", userID).
		Where("transaction_date >= ?", startDate).
		Where("transaction_date <= ?", endDate).
		Limit(limit).
		Offset(offset).
		Order("transaction_date DESC, id DESC").
		Find(&transactions).
		Error
	
	if err != nil {
		return nil, err
	}

	return transactions, nil
}

func (r *TransactionRepository) SumTransactionsByUserAndDateRange(
	ctx context.Context, 
	userID uint64,
	transactionType domain.TransactionType, 
	startDate time.Time, 
	endDate time.Time,
) (*int64, error) {
	var total int64

	err := r.db.
		WithContext(ctx).
		Model(&domain.Transaction{}).
		Where("user_id = ?", userID).
		Where("type = ?", transactionType).
		Where("transaction_date >= ?", startDate).
		Where("transaction_date <= ?", endDate).
		Select("COALSCE(SUM(amount), 0)").
		Scan(&total).
		Error
	
	if err != nil {
		return nil, err
	}

	return &total, nil
}

