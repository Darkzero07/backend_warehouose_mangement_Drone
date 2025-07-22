package models

import "gorm.io/gorm"

type Warranty struct {
	gorm.Model
	DroneID           uint   `gorm:"not null"`
	SerialNumber      string `gorm:"not null"`
	BuyDate           string `gorm:"not null"`
	TimeWarranty      string `gorm:"not null"`
	Status            string `gorm:"not null"`
	BoxID            uint `gorm:"not null"`
	Lot			  string 
	Remark           string 
}