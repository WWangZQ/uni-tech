package model

import (
	"time"
)

type Activity struct {
	ID               int64     `gorm:"primaryKey;autoIncrement" json:"id"`
	Title            string    `gorm:"size:255;not null" json:"title"`
	Description      string    `gorm:"type:text" json:"description"`
	Organizer        string    `gorm:"size:100" json:"organizer"`
	Speaker          string    `gorm:"size:100" json:"speaker"`
	Location         string    `gorm:"size:255" json:"location"`
	ActivityType     string    `gorm:"type:enum('lecture','concert','competition','exhibition','other');not null" json:"activity_type"`
	StartTime        time.Time `gorm:"not null" json:"start_time"`
	EndTime          time.Time `gorm:"not null" json:"end_time"`
	TotalTickets     int       `gorm:"not null" json:"total_tickets"`
	RemainingTickets int       `gorm:"not null" json:"remaining_tickets"`
	Price            float64   `gorm:"type:decimal(10,2);default:0.00" json:"price"`
	SeckillStart     *time.Time `gorm:"index" json:"seckill_start,omitempty"`
	SeckillEnd       *time.Time `gorm:"index" json:"seckill_end,omitempty"`
	Status           string    `gorm:"type:enum('draft','seckill','ongoing','ended','cancelled');default:'draft'" json:"status"`
	CoverImage       string    `gorm:"size:500" json:"cover_image"`
	ViewCount        int       `gorm:"default:0" json:"view_count"`
	CreatedAt        time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt        time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}

func (Activity) TableName() string {
	return "activities"
}
