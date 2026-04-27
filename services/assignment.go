package services

import (
	"errors"
	"gin-quickstart/database"
	"gin-quickstart/models"
)

func FindLeastBusyDoctor(examinationID uint) (uint, error) {
	var doctors []models.Doctor

	err := database.DB.Joins("JOIN doctor_examinations ON doctor_examinations.doctor_id = doctors.id").
		Where("doctor_examinations.examination_id = ?", examinationID).
		Find(&doctors).Error

	if err != nil || len(doctors) == 0 {
		return 0, errors.New("no qualified doctors found for this examination")
	}

	var bestDoctorID uint
	minAppointments := int64(-1)

	for _, doctor := range doctors {
		var count int64
		database.DB.Model(&models.Appointment{}).
			Where("doctor_id = ? AND is_finished = ?", doctor.ID, false).
			Count(&count)

		if minAppointments == -1 || count < minAppointments {
			minAppointments = count
			bestDoctorID = doctor.ID
		}
	}

	return bestDoctorID, nil
}
