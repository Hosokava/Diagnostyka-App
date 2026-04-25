package models

import (
	"time"

	"gorm.io/gorm"
)

type Appointment struct {
	gorm.Model
	ExaminationID   uint        `gorm:"not null"`
	Examination     Examination `gorm:"constraint:OnUpdate:CASCADE,OnDelete:SET NULL;"`
	DoctorID        uint        `gorm:"not null"`
	Doctor          Doctor      `gorm:"constraint:OnUpdate:CASCADE,OnDelete:SET NULL;"`
	PatientID       uint        `gorm:"not null"`
	Patient         Patient     `gorm:"constraint:OnUpdate:CASCADE,OnDelete:SET NULL;"`
	IsFinished      bool        `gorm:"default:false"`
	QRCodeHash      string      `gorm:"type:varchar(255)"`
	Date            time.Time   `gorm:"not null"`
	Result          string      `gorm:"type:text"`
	DiagnosticNotes string      `gorm:"type:text"`
}
