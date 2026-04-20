package models

import (
	"gorm.io/gorm"
)

type Doctor struct {
	gorm.Model
	Email          string `gorm:"type:varchar(255);uniqueIndex;not null"`
	PasswordHash   string `gorm:"type:varchar(255);not null"`
	Specialization string `gorm:"type:varchar(100)"`
}
