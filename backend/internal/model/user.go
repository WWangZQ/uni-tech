package model

import (
	"time"
)

type User struct {
	ID           int64     `gorm:"primaryKey;autoIncrement" json:"id"`
	StudentID    string    `gorm:"uniqueIndex;size:20;not null" json:"student_id"`
	Name         string    `gorm:"size:100;not null" json:"name"`
	Email        string    `gorm:"uniqueIndex;size:255" json:"email"`
	Phone        string    `gorm:"size:20" json:"phone"`
	Role         string    `gorm:"type:enum('undergraduate','postgraduate','teacher','admin');default:'undergraduate'" json:"role"`
	CreditScore  int       `gorm:"default:100" json:"credit_score"`
	NoShowCount  int       `gorm:"default:0" json:"no_show_count"`
	QuotaHours   int       `gorm:"default:10" json:"quota_hours"`
	Status       string    `gorm:"type:enum('active','suspended','deleted');default:'active'" json:"status"`
	PasswordHash string    `gorm:"size:255;not null" json:"-"`
	CreatedAt    time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt    time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}

func (User) TableName() string {
	return "users"
}
