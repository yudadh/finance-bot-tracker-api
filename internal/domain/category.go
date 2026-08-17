package domain

import (
	"time"

	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type CategoryType string

const (
	CategoryTypeIncome  CategoryType = "income"
	CategoryTypeExpense CategoryType = "expense"
)

type Category struct {
	ID           uint64                      `gorm:"column:id;primaryKey"`
	Name         string                      `gorm:"column:name"`
	Type         CategoryType                `gorm:"column:type"`
	Keywords     datatypes.JSONSlice[string] `gorm:"column:keywords;type:json"`
	IsDefault    bool                        `gorm:"column:is_default"`
	CreatedAt    time.Time                   `gorm:"column:created_at;autoCreateTime"`
	UpdatedAt    time.Time                   `gorm:"column:updated_at;autoCreateTime;autoUpdateTime"`
	DeletedAt    gorm.DeletedAt              `gorm:"column:deleted_at"`
	Transactions []Transaction               `gorm:"foreignKey:CategoryID;references:ID"`
}

func (c *Category) TableName() string {
	return "categories"
}