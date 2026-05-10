package model

import (
	"time"
)

type OrderStatus string

const (
	OrderStatusPending   OrderStatus = "pending"
	OrderStatusConfirmed OrderStatus = "confirmed"
	OrderStatusPaid     OrderStatus = "paid"
	OrderStatusCancelled OrderStatus = "cancelled"
	OrderStatusNoShow   OrderStatus = "no_show"
	OrderStatusCompleted OrderStatus = "completed"
)

type Order struct {
	ID               int64       `gorm:"primaryKey;autoIncrement" json:"id"`
	OrderNo          string      `gorm:"uniqueIndex;size:32;not null" json:"order_no"`
	UserID           int64       `gorm:"not null;index" json:"user_id"`
	OrderType        string      `gorm:"type:enum('space','activity');not null" json:"order_type"`
	Status           OrderStatus `gorm:"type:enum('pending','confirmed','paid','cancelled','no_show','completed');default:'pending';index" json:"status"`
	TotalAmount      float64     `gorm:"type:decimal(10,2);default:0.00" json:"total_amount"`
	CreditDeduction  int         `gorm:"default:0" json:"credit_deduction"`
	PaymentDeadline  *time.Time  `json:"payment_deadline,omitempty"`
	PaidAt           *time.Time  `json:"paid_at,omitempty"`
	CancelledAt      *time.Time  `json:"cancelled_at,omitempty"`
	CancelReason     string      `gorm:"size:255" json:"cancel_reason,omitempty"`
	Version          int         `gorm:"default:0" json:"version"`
	Remark           string      `gorm:"type:text" json:"remark,omitempty"`
	CreatedAt        time.Time   `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt        time.Time   `gorm:"autoUpdateTime" json:"updated_at"`

	Items []OrderItem `gorm:"foreignKey:OrderID" json:"items,omitempty"`
}

func (Order) TableName() string {
	return "orders"
}

type OrderItem struct {
	ID          int64     `gorm:"primaryKey;autoIncrement" json:"id"`
	OrderID     int64     `gorm:"not null;index" json:"order_id"`
	ResourceID  *int64    `gorm:"index" json:"resource_id,omitempty"`
	ActivityID  *int64    `gorm:"index" json:"activity_id,omitempty"`
	TicketCount int       `gorm:"default:1" json:"ticket_count"`
	UnitPrice   float64   `gorm:"type:decimal(10,2);default:0.00" json:"unit_price"`
	SlotDate    *string   `gorm:"type:date" json:"slot_date,omitempty"`
	StartTime   *string   `gorm:"type:time" json:"start_time,omitempty"`
	EndTime     *string   `gorm:"type:time" json:"end_time,omitempty"`
	CreatedAt   time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt   time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}

func (OrderItem) TableName() string {
	return "order_items"
}
