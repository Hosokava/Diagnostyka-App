package database

import (
	"gin-quickstart/models"
	"log"

	"golang.org/x/crypto/bcrypt"
)

func SeedExaminations() {
	var count int64
	DB.Model(&models.Examination{}).Count(&count)
	if count > 0 {
		log.Println("Database already seeded with examinations.")
		return
	}

	exams := []models.Examination{
		{Name: "Full Blood Count", Description: "Basic blood test assessing overall health.", Price: 45.00},
		{Name: "Urinalysis", Description: "Assessment of physical and chemical parameters of urine.", Price: 25.00},
		{Name: "Abdominal Ultrasound", Description: "Imaging of internal organs.", Price: 150.00},
		{Name: "Chest X-Ray", Description: "X-ray of the chest (lungs and heart).", Price: 100.00},
		{Name: "Resting ECG", Description: "Recording of the electrical activity of the heart.", Price: 60.00},
		{Name: "Glucose Level", Description: "Measurement of sugar concentration in the blood.", Price: 15.00},
		{Name: "Thyroid Panel", Description: "Diagnostics of thyroid diseases.", Price: 120.00},
		{Name: "Lipid Profile", Description: "Blood lipid profile (Cholesterol, LDL, HDL, TG).", Price: 80.00},
		{Name: "Echocardiogram", Description: "Ultrasound examination of the heart.", Price: 200.00},
		{Name: "Creatinine Test", Description: "Assessment of kidney function.", Price: 20.00},
	}

	if err := DB.Create(&exams).Error; err != nil {
		log.Printf("Failed to seed examinations: %v", err)
	} else {
		log.Println("Successfully seeded 10 examinations.")
	}

	SeedDoctors()
}

func SeedDoctors() {
	var count int64
	DB.Model(&models.Doctor{}).Count(&count)
	if count > 0 {
		log.Println("Database already seeded with doctors.")
		return
	}

	var exams []models.Examination
	DB.Find(&exams)

	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
	pass := string(hashedPassword)

	doctors := []models.Doctor{
		{
			FirstName: "Gregory", LastName: "House", Specialization: "Diagnostics",
			Email: "house@med.pl", PasswordHash: pass,
			Examinations: []models.Examination{exams[0], exams[1], exams[2], exams[3], exams[4]},
		},
		{
			FirstName: "James", LastName: "Wilson", Specialization: "Oncology",
			Email: "wilson@med.pl", PasswordHash: pass,
			Examinations: []models.Examination{exams[5], exams[6], exams[7], exams[8], exams[9]},
		},
		{
			FirstName: "Lisa", LastName: "Cuddy", Specialization: "Endocrinology",
			Email: "cuddy@med.pl", PasswordHash: pass,
			Examinations: []models.Examination{exams[0], exams[2], exams[4], exams[6], exams[8]},
		},
		{
			FirstName: "Eric", LastName: "Foreman", Specialization: "Neurology",
			Email: "foreman@med.pl", PasswordHash: pass,
			Examinations: []models.Examination{exams[1], exams[3], exams[5], exams[7], exams[9]},
		},
	}

	if err := DB.Create(&doctors).Error; err != nil {
		log.Printf("Failed to seed doctors: %v", err)
	} else {
		log.Println("Successfully seeded 4 doctors with English data and 'password123' as password.")
	}
}
