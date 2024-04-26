package users

import (
	"fmt"
	"time"

	"github.com/dsseng/wiso/pkg/radius"
	"gorm.io/gorm"
)

type DeviceSession struct {
	gorm.Model
	UserID     uint
	DueDate    time.Time
	RadcheckID uint
	Inactive   bool
	MAC        string
}

func (s *DeviceSession) BeforeDelete(tx *gorm.DB) error {
	tx.Delete(&radius.RadCheck{}, s.RadcheckID)
	return nil
}

func CleanupOutdatedSessions(db *gorm.DB) error {
	var sess []DeviceSession
	res := db.Where("inactive = ?", false).Find(&sess, "due_date < ?", time.Now())
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return nil
	}

	radchecks := []uint{}
	for i := range sess {
		sess[i].Inactive = true
		radchecks = append(radchecks, sess[i].RadcheckID)
	}

	res = db.Delete(&[]radius.RadCheck{}, radchecks)
	if res.Error != nil {
		return res.Error
	}

	res = db.Save(&sess)
	if res.Error != nil {
		return res.Error
	}

	fmt.Printf("Cleaned up %v outdated sessions\n", res.RowsAffected)
	return nil
}

// TODO: handle groups
func StartSession(db *gorm.DB, user User, mac string, dueDate time.Time) error {
	radcheck := radius.RadCheck{
		Username:  mac,
		Attribute: "Cleartext-Password",
		Op:        ":=",
		Value:     "macauth",
	}
	res := db.Create(&radcheck)
	if res.Error != nil {
		return res.Error
	}

	sess := DeviceSession{
		DueDate:    dueDate,
		RadcheckID: radcheck.ID,
		MAC:        mac,
	}

	user.DeviceSessions = append(
		user.DeviceSessions,
		sess,
	)
	db.Save(user)

	return nil
}

type User struct {
	gorm.Model
	Username       string
	Picture        string
	FullName       string
	DeviceSessions []DeviceSession
}

func FindSingle(db *gorm.DB, username string) ([]User, error) {
	user := []User{}
	res := db.Limit(1).Find(&user, "username = ?", username)
	if res.Error != nil {
		return nil, res.Error
	}
	return user, nil
}
