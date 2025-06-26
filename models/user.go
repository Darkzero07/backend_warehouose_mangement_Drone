package models

import (
	"gorm.io/gorm"
	"time"
)

type User struct {
	gorm.Model
	Username string `gorm:"unique;not null"`
	Password string `gorm:"not null"`
	Role     string `gorm:"not null;default:'user'"`
	// Email 			 string `gorm:"unique;not null"`
	ResetToken       string
	ResetTokenExpiry time.Time
}
