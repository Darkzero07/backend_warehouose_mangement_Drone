package models

import "gorm.io/gorm"

type DamageReport struct {
	gorm.Model
	ItemID      uint   `gorm:"not null"`
	Item        Item   `gorm:"foreignkey:ItemID"` //Belongs To relationship
	ReporterID  uint   `gorm:"not null"`
	Reporter    User   `gorm:"foreignkey:ReporterID"`
	ProjectID   uint   `gorm:"not null"`
	Project     Project `gorm:"foreignkey:ProjectID"` //Belongs To relationship
	Description string `gorm:"not null"`
	Status      string `gorm:"default:'Pending'"` 
	Broken_Drone uint  
}