package domain

import "time"

type UserStatus string

const (
	UserStatusActive   UserStatus = "active"
	UserStatusInactive UserStatus = "inactive"
	UserStatusBlocked  UserStatus = "blocked"
)

type User struct {
	ID               uint64        `gorm:"column:id;primaryKey"`
	TelegramID       int64         `gorm:"column:telegram_id;unique"`
	TelegramUsername string        `gorm:"column:telegram_username"`
	FirstName        string        `gorm:"column:first_name"`
	LastName         string        `gorm:"column:last_name"`
	LanguageCode     string        `gorm:"column:language_code"`
	Timezone         string        `gorm:"column:timezone"`
	Status           UserStatus    `gorm:"column:status"`
	LastSeenAt       *time.Time    `gorm:"column:last_seen_at"`
	CreatedAt        time.Time     `gorm:"column:created_at;autoCreateTime"`
	UpdatedAt        time.Time     `gorm:"column:updated_at;autoCreateTime;autoUpdateTime"`
	Budgets          []Budget      `gorm:"foreignKey:UserID;references:ID"`
	Transactions     []Transaction `gorm:"foreignKey:UserID;references:ID"`
	ActivityLogs     []ActivityLog `gorm:"foreignKey:UserID;references:ID"`
}

func (u *User) TableName() string {
	return "users"
}
