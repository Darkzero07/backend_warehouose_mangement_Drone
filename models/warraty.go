// models/transactionBorrow.go
package models

import "gorm.io/gorm"

type Warranty struct {
	gorm.Model
	ItemID            uint   `gorm:"not null"`
	Item              Item   `gorm:"foreignkey:ItemID"`
	BuyDate           string `gorm:"not null"`
	Description       string `gorm:"not null"`
	TimeWarranty      string `gorm:"not null"`
	RemainingWarranty string `gorm:"not null"`
	SerialNumber      string `gorm:"not null"`
	ItemStatus        string `gorm:"-"`
}

 
// Method to load warranty with item status
func (w *Warranty) LoadWithStatus(db *gorm.DB, id uint) error {
	if err := db.Preload("Item").First(w, id).Error; err != nil {
		return err
	}
	w.ItemStatus = w.Item.Status
	return nil
}