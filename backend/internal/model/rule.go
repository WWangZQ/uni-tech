package model

import (
	"time"
)

type BookingRule struct {
	ID          int64      `gorm:"primaryKey;autoIncrement" json:"id"`
	RuleCode    string     `gorm:"uniqueIndex;size:50;not null" json:"rule_code"`
	RuleName    string     `gorm:"size:100;not null" json:"rule_name"`
	RuleType    string     `gorm:"type:enum('quota','credit','duration','blacklist','custom');not null" json:"rule_type"`
	RuleConfig  JSONMap    `gorm:"type:json;not null" json:"rule_config"`
	Priority    int        `gorm:"default:0" json:"priority"`
	Status      string     `gorm:"type:enum('enabled','disabled');default:'enabled'" json:"status"`
	EffectStart *time.Time `json:"effect_start,omitempty"`
	EffectEnd   *time.Time `json:"effect_end,omitempty"`
	CreatedAt   time.Time  `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt   time.Time  `gorm:"autoUpdateTime" json:"updated_at"`
}

func (BookingRule) TableName() string {
	return "booking_rules"
}
