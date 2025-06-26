// models/transactionReturn.go
package models

import "gorm.io/gorm"

type TransactionReturn struct {
	gorm.Model
	UserID         uint              `gorm:"not null"`
	User           User              `gorm:"foreignkey:UserID"`
	ItemID         uint              `gorm:"not null"`
	Item           Item              `gorm:"foreignkey:ItemID"`
	ProjectID      uint              `gorm:"not null"`
	Project        Project           `gorm:"foreignkey:ProjectID"`
	ReturnQuantity int               `gorm:"not null"`
	ReturnDate     string            `gorm:"not null"`
	BorrowID       uint              `gorm:"not null"` // Reference to the original borrow transaction
	Borrow         TransactionBorrow `gorm:"foreignkey:BorrowID"`
}
