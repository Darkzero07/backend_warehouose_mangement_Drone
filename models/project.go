package models

import (
	"gorm.io/gorm"
	"time"
)

type Project struct {
	gorm.Model
	Name            string `gorm:"unique;not null"`
	Description     string
	StartDate       string
	EndDate         string
	Number_of_Drone uint
	Location        string
	CreatedAt       time.Time `gorm:"type:timestamp"`
	UpdatedAt       time.Time `gorm:"type:timestamp"`
}
