package domain

import (
	"time"

	"gorm.io/datatypes"
)

type ActivityLog struct {
	ID          uint64          `gorm:"column:id;primaryKey"`
	UserID      *uint64         `gorm:"column:user_id"`
	AdminUserID *uint64         `gorm:"column:admin_user_id"`
	Action      string          `gorm:"column:action"`
	EntityType  string          `gorm:"column:entity_type"`
	EntityID    *uint64         `gorm:"column:entity_id"`
	Metadata    *datatypes.JSON `gorm:"column:metadata;type:json"`
	IpAddress   *string         `gorm:"column:ip_address"`
	UserAgent   *string         `gorm:"column:user_agent"`
	CreatedAt   time.Time       `gorm:"column:created_at;autoCreateTime"`
	User        *User           `gorm:"foreignKey:UserID;references:ID"`
	AdminUser   *AdminUser      `gorm:"foreignKey:AdminUserID;references:ID"`
}

func (a *ActivityLog) TableName() string {
	return "activity_logs"
}
