package database

import (
	"time"

	"github.com/yudadh/finance-bot-tracker-api/internal/config"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func NewGormDB(config config.DatabaseConfig) (*gorm.DB, error) {
	dsn := config.GormDSN()
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		TranslateError: true,
	})

	if err != nil {
		return nil, err
	}

	sqlDb, err := db.DB()

	sqlDb.SetMaxOpenConns(10)
	sqlDb.SetMaxIdleConns(5)
	sqlDb.SetConnMaxLifetime(30 * time.Minute)
	sqlDb.SetConnMaxIdleTime(5 * time.Minute)

	return db, nil
}