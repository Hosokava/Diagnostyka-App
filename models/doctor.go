package models

import (
	"gorm.io/gorm"
)

type Doctor struct {
	gorm.Model
	Email          string        `gorm:"type:varchar(255);uniqueIndex;not null"`
	PasswordHash   string        `gorm:"type:varchar(255);not null" json:"-"`
	FirstName      string        `gorm:"type:varchar(100)"`
	LastName       string        `gorm:"type:varchar(100)"`
	Specialization string        `gorm:"type:varchar(100)"`
	Examinations   []Examination `gorm:"many2many:doctor_examinations;"`
}
