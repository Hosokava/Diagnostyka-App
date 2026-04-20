package models

import (
	"time"

	"gorm.io/gorm"
)

type RefreshToken struct {
	gorm.Model
	TokenHash string    `gorm:"type:varchar(255);uniqueIndex;not null"`
	UserID    uint      `gorm:"not null"`
	UserType  string    `gorm:"type:varchar(20);not null"`
	ExpiresAt time.Time `gorm:"not null"`
}
