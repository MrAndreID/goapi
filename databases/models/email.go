package models

import (
	"time"

	"gorm.io/gorm"
)

type Email struct {
	ID        string         `gorm:"primaryKey;Column:id;type:varchar(45)" json:"id"`
	CreatedAt time.Time      `gorm:"Column:created_at;type:timestamptz;not null" json:"createdAt"`
	UpdatedAt time.Time      `gorm:"Column:updated_at;type:timestamptz;not null" json:"updatedAt"`
	DeletedAt gorm.DeletedAt `gorm:"Column:deleted_at;type:timestamptz" json:"deletedAt"`
	UserID    string         `gorm:"Column:user_id;type:varchar(45);not null" json:"userId"`
	Email     string         `gorm:"Column:email;type:varchar(255);not null" json:"email"`
}

func (Email) TableName() string {
	return "emails"
}
