// models/transactionBorrow.go
package models

import "gorm.io/gorm"

type TransactionBorrow struct {
	gorm.Model
	UserID     uint    `gorm:"not null"`
	User       User    `gorm:"foreignkey:UserID"`
	ItemID     uint    `gorm:"not null"`
	Item       Item    `gorm:"foreignkey:ItemID"`
	ProjectID  uint    `gorm:"not null"`
	Project    Project `gorm:"foreignkey:ProjectID"`
	BorrowQuantity   int     `gorm:"not null"`
	BorrowDate string  `gorm:"not null"`
	DueDate    string  `gorm:"not null"` // Expected return date
}