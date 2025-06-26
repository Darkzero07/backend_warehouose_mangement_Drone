package models

import "gorm.io/gorm"

type AuditLog struct {
	gorm.Model
	UserID    uint   `gorm:"not null"`
	User      User   `gorm:"foreignkey:UserID"`
	Action    string `gorm:"not null"` // e.g., "create_item", "update_item"
	TableName string `gorm:"not null"` // e.g., "items"
	RecordID  uint   `gorm:"not null"`
	OldValue  string `gorm:"type:text"`
	NewValue  string `gorm:"type:text"`
	IPAddress string
}
