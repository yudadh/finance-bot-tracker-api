package database

import (
	"time"

	"github.com/yudadh/finance-bot-tracker-api/internal/config"
	"github.com/yudadh/finance-bot-tracker-api/internal/utils"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func NewGormDB(config config.DatabaseConfig) *gorm.DB {
	dsn := config.GormDSN()
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	utils.PanicIfError(err)

	sqlDb, err := db.DB()
	utils.PanicIfError(err)

	sqlDb.SetMaxOpenConns(10)
	sqlDb.SetMaxIdleConns(5)
	sqlDb.SetConnMaxLifetime(30 * time.Minute)
	sqlDb.SetConnMaxIdleTime(5 * time.Minute)

	return db
}