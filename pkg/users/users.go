package users

import (
	"time"

	"github.com/dsseng/wiso/pkg/radius"
	"gorm.io/gorm"
)

type DeviceSession struct {
	gorm.Model
	UserID     uint
	DueDate    time.Time
	RadcheckID uint
}

func (s *DeviceSession) BeforeDelete(tx *gorm.DB) error {
	tx.Delete(&radius.RadCheck{}, s.RadcheckID)
	return nil
}

type User struct {
	gorm.Model
	Username       string
	Picture        string
	FullName       string
	DeviceSessions []DeviceSession
}
