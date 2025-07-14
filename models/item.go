package models

import "gorm.io/gorm"

type Item struct {
	gorm.Model
	Name        string `gorm:"not null"`
	Description string
	Quantity    int      `gorm:"not null"`
	Status      string   
	CategoryID  uint     // Foreign key to Category table
	Category    Category // Belongs To relationship
	Remark      string
}
