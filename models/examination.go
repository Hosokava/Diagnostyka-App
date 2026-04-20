package models

import (
	"gorm.io/gorm"
)

type Examination struct {
	gorm.Model
	Name        string  `gorm:"type:varchar(100);not null"`
	Description string  `gorm:"type:text"`
	Price       float64 `gorm:"type:decimal(10,2);not null"`
}
