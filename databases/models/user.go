package models

import (
	"time"

	"gorm.io/gorm"
)

type User struct {
	ID        string         `gorm:"primaryKey;Column:id;type:varchar(45)" json:"id"`
	CreatedAt time.Time      `gorm:"Column:created_at;type:timestamptz;not null" json:"createdAt"`
	UpdatedAt time.Time      `gorm:"Column:updated_at;type:timestamptz;not null" json:"updatedAt"`
	DeletedAt gorm.DeletedAt `gorm:"Column:deleted_at;type:timestamptz" json:"deletedAt"`
	Name      string         `gorm:"Column:name;type:varchar(255);not null" json:"name"`
	Emails    []Email        `gorm:"foreignKey:UserID;references:ID" json:"emails"`
}

func (User) TableName() string {
	return "users"
}
