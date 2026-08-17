package domain

import (
	"time"

	"gorm.io/datatypes"
)

type ParserType string

const (
	ParserTypeRuleBased ParserType = "rule_based"
	ParserTypeAI        ParserType = "ai"
)

type ParserAttempt struct {
	ID            uint64          `gorm:"column:id;primaryKey"`
	UserID        uint64          `gorm:"column:user_id"`
	RawText       string          `gorm:"column:raw_text"`
	ParserType    ParserType      `gorm:"column:parser_type"`
	Success       bool            `gorm:"column:success"`
	Confidence    *float64        `gorm:"column:confidence"`
	ParsedPayload *datatypes.JSON `gorm:"column:parser_payload;type:json"`
	ErrorMessage  *string         `gorm:"column:error_message"`
	CreatedAt     time.Time       `gorm:"column:created_at;autoCreateTime"`
	User          User            `gorm:"foreignKey:UserID;references:ID"`
}

func (p *ParserAttempt) TableName() string {
	return "parser_attempts"
}
