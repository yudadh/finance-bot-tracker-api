package domain

import "time"

type AdminUserStatus string

const (
	AdminUserStatusActive   AdminUserStatus = "active"
	AdminUserStatusInactive AdminUserStatus = "inactive"
)

type AdminUser struct {
	ID           uint64          `gorm:"column:id;primaryKey"`
	Email        string          `gorm:"column:email"`
	PasswordHash string          `gorm:"column:password_hash"`
	Name         string          `gorm:"column:name"`
	Status       AdminUserStatus `gorm:"column:status"`
	LastLoginAt  *time.Time       `gorm:"column:last_login_at"`
	CreatedAt    time.Time       `gorm:"column:created_at;autoCreateTime"`
	UpdatedAt    time.Time       `gorm:"column:updated_at;autoCreateTime;autoUpdateTime"`
	ActivityLogs []ActivityLog   `gorm:"foreignKey:AdminUserID;references:ID"`
}

func (a *AdminUser) TableName() string {
	return "admin_users"
}