package models

import "gorm.io/gorm"

type Transaction struct {
	gorm.Model
	UserID    uint    `gorm:"not null"`
	User      User    `gorm:"foreignkey:UserID"`
	ItemID    uint    `gorm:"not null"`
	Item      Item    `gorm:"foreignkey:ItemID"`
	ProjectID uint    `gorm:"not null"`
	Project   Project `gorm:"foreignkey:ProjectID"`
	Quantity  int     `gorm:"not null"`
	Type      string  `gorm:"not null"` // "borrow" or "return"
	Status    string  // e.g., "Pending", "Approved", "Rejected"
	BorrowDate string `gorm:"not null"`
	ReturnDate string `gorm:"not null"`
}