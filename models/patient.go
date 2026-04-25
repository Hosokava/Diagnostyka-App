package models

import (
	"gorm.io/gorm"
)

type Patient struct {
	gorm.Model
	Email        string `gorm:"type:varchar(255);uniqueIndex;not null"`
	PasswordHash string `gorm:"type:varchar(255);not null" json:"-"`
	FirstName    string `gorm:"type:varchar(100)"`
	LastName     string `gorm:"type:varchar(100)"`
	PESEL        string `gorm:"type:varchar(11)"`
}
