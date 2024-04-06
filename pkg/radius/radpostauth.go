package radius

import "time"

type RadPostAuth struct {
	ID               uint64 `gorm:"primary_key;type:bigint"`
	Username         string // NOT NULL
	Pass             string
	Reply            string
	CalledStationId  string    `gorm:"column:calledstationid"`
	CallingStationId string    `gorm:"column:callingstationid"`
	AuthDate         time.Time `gorm:"column:authdate"` // NOT NULL
	Class            string
}

func (RadPostAuth) TableName() string {
	return "radpostauth"
}
