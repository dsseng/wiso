package radius

type RadCheck struct {
	ID        uint `gorm:"primary_key"`
	Username  string
	Attribute string
	Op        string `gorm:"type:varchar(2)"`
	Value     string
}

func (RadCheck) TableName() string {
	return "radcheck"
}
