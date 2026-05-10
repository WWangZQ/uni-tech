package model

import (
	"time"
)

type TimeSlot struct {
	ID          int64     `gorm:"primaryKey;autoIncrement" json:"id"`
	ResourceID  int64     `gorm:"not null;index" json:"resource_id"`
	SlotDate    time.Time `gorm:"type:date;not null;index" json:"slot_date"`
	StartTime   string    `gorm:"type:time;not null" json:"start_time"`
	EndTime     string    `gorm:"type:time;not null" json:"end_time"`
	BufferStart string    `gorm:"type:time;not null" json:"buffer_start"`
	BufferEnd   string    `gorm:"type:time;not null" json:"buffer_end"`
	Status      string    `gorm:"type:enum('available','locked','booked','cancelled');default:'available';index" json:"status"`
	BookingID   *int64    `gorm:"index" json:"booking_id,omitempty"`
	Version     int       `gorm:"default:0" json:"version"`
	CreatedAt   time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt   time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}

func (TimeSlot) TableName() string {
	return "time_slots"
}
