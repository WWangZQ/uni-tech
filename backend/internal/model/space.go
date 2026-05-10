package model

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"
)

type ResourceType string

const (
	ResourceTypeAcademic ResourceType = "academic"
	ResourceTypeSports  ResourceType = "sports"
)

type Resource struct {
	ID          int64        `gorm:"primaryKey;autoIncrement" json:"id"`
	Type        ResourceType `gorm:"type:enum('academic','sports');not null;index" json:"type"`
	Name        string       `gorm:"size:255;not null" json:"name"`
	Code        string       `gorm:"uniqueIndex;size:50;not null" json:"code"`
	Capacity    int          `gorm:"not null" json:"capacity"`
	Location    string       `gorm:"size:255" json:"location"`
	Floor       int          `json:"floor"`
	Building    string       `gorm:"size:100" json:"building"`
	ImageURL    string       `gorm:"size:500" json:"image_url"`
	Description string       `gorm:"type:text" json:"description"`
	Status      string       `gorm:"type:enum('active','inactive','maintenance');default:'active';index" json:"status"`
	CreatedAt   time.Time    `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt   time.Time    `gorm:"autoUpdateTime" json:"updated_at"`

	AcademicSpace *AcademicSpace  `gorm:"foreignKey:ResourceID" json:"academic_space,omitempty"`
	SportsFacility *SportsFacility `gorm:"foreignKey:ResourceID" json:"sports_facility,omitempty"`
}

func (Resource) TableName() string {
	return "resources"
}

type AcademicSpace struct {
	ID             int64           `gorm:"primaryKey;autoIncrement" json:"id"`
	ResourceID     int64           `gorm:"uniqueIndex;not null" json:"resource_id"`
	BufferMinutes  int             `gorm:"default:5" json:"buffer_minutes"`
	MinDuration    int             `gorm:"default:30" json:"min_duration"`
	MaxDuration    int             `gorm:"default:240" json:"max_duration"`
	AllowRecurring bool            `gorm:"default:false" json:"allow_recurring"`
	Equipment      JSONMap         `gorm:"type:json" json:"equipment"`
	CreatedAt      time.Time       `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt      time.Time       `gorm:"autoUpdateTime" json:"updated_at"`
}

func (AcademicSpace) TableName() string {
	return "academic_spaces"
}

type SportsFacility struct {
	ID           int64     `gorm:"primaryKey;autoIncrement" json:"id"`
	ResourceID   int64     `gorm:"uniqueIndex;not null" json:"resource_id"`
	SlotDuration int       `gorm:"default:60" json:"slot_duration"`
	Combinable   bool      `gorm:"default:true" json:"combinable"`
	CourtCount   int       `gorm:"default:1" json:"court_count"`
	SportType    string    `gorm:"type:enum('basketball','tennis','badminton','football','pingpong','other');not null" json:"sport_type"`
	Indoor       bool      `gorm:"default:true" json:"indoor"`
	CreatedAt    time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt    time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}

func (SportsFacility) TableName() string {
	return "sports_facilities"
}

// JSONMap 自定义类型，用于处理 JSON 类型的 equipment 字段
type JSONMap map[string]interface{}

func (j JSONMap) Value() (driver.Value, error) {
	if j == nil {
		return nil, nil
	}
	return json.Marshal(j)
}

func (j *JSONMap) Scan(value interface{}) error {
	if value == nil {
		*j = nil
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}
	return json.Unmarshal(bytes, j)
}
