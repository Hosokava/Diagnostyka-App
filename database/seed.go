package database

import (
	"gin-quickstart/models"
	"log"
)

func SeedExaminations() {
	var count int64
	DB.Model(&models.Examination{}).Count(&count)
	if count > 0 {
		log.Println("Database already seeded with examinations.")
		return
	}

	exams := []models.Examination{
		{Name: "Morfologia krwi", Description: "Podstawowe badanie krwi oceniające ogólny stan zdrowia.", Price: 45.00},
		{Name: "Badanie ogólne moczu", Description: "Ocena parametrów fizykochemicznych moczu.", Price: 25.00},
		{Name: "USG jamy brzusznej", Description: "Badanie obrazowe narządów wewnętrznych.", Price: 150.00},
		{Name: "RTG klatki piersiowej", Description: "Prześwietlenie klatki piersiowej (płuc i serca).", Price: 100.00},
		{Name: "EKG spoczynkowe", Description: "Zapis czynności elektrycznej serca.", Price: 60.00},
		{Name: "Poziom glukozy", Description: "Pomiar stężenia cukru we krwi.", Price: 15.00},
		{Name: "Panel tarczycowy (TSH, FT3, FT4)", Description: "Diagnostyka chorób tarczycy.", Price: 120.00},
		{Name: "Lipidogram (Cholesterol, LDL, HDL, TG)", Description: "Profil lipidowy kwi.", Price: 80.00},
		{Name: "Echo serca", Description: "Badanie echokardiograficzne serca.", Price: 200.00},
		{Name: "Kreatynina w surowicy", Description: "Ocena sprawności nerek.", Price: 20.00},
	}

	if err := DB.Create(&exams).Error; err != nil {
		log.Printf("Failed to seed examinations: %v", err)
	} else {
		log.Println("Successfully seeded 10 examinations.")
	}
}
