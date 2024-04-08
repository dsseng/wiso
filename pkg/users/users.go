package users

import (
	"time"

	"gorm.io/gorm"
)

type DeviceSession struct {
	gorm.Model
	ID         uint64 `gorm:"primary_key;unique"`
	UserID     uint
	DueDate    time.Time
	RadcheckID uint
}

type User struct {
	gorm.Model
	ID             uint64 `gorm:"primary_key;unique"`
	Username       string
	Picture        string
	FullName       string
	DeviceSessions []DeviceSession
}
