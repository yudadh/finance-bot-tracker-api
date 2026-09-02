package config

import (
	"fmt"
	"time"

	"github.com/spf13/viper"
)

func NewConfig() viper.Viper {
	var config viper.Viper = *viper.New()
	config.SetConfigFile(".env")

	return config
}

type Config struct {
	App AppConfig
	Database DatabaseConfig
}

func New(cfg viper.Viper) Config {
	return Config{
		App: NewAppConfig(cfg),
		Database: NewDatabaseConfig(cfg),
	}
}

type DatabaseConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	Name     string
}

func NewDatabaseConfig(config viper.Viper) DatabaseConfig {
	return DatabaseConfig{
		Host:     config.GetString("DATABASE_HOST"),
		Port:     config.GetString("DATABASE_PORT"),
		User:     config.GetString("DATABASE_USER"),
		Password: config.GetString("DATABASE_PASSWORD"),
		Name:     config.GetString("DATABASE_NAME"),
	}
}

func (c *DatabaseConfig) GormDSN() string {
	return fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=UTC",
		c.User,
		c.Password,
		c.Host,
		c.Port,
		c.Name,
	)
}

type JWTConfig struct {
	Secret string
	Issuer string
	TTL time.Duration
}

type AppConfig struct {
	AppEnv string
	AppPort string
	AppName string
	TelegramBotToken string
	JWT JWTConfig
}

func NewAppConfig(config viper.Viper) AppConfig {
	return AppConfig{
		AppEnv: config.GetString("APP_ENV"),
		AppPort: config.GetString("APP_PORT"),
		AppName: config.GetString("APP_NAME"),
		TelegramBotToken: config.GetString("TELEGRAM_BOT_TOKEN"),
		JWT: JWTConfig{
			Secret: config.GetString("JWT_SECRET"),
			Issuer: config.GetString("JWT_ISSUER"),
			TTL: config.GetDuration("JWT_TTL"),
		},
	}
}

func (c *AppConfig) HTTPAddress() string {
	return fmt.Sprintf(":%s", c.AppPort)
}
