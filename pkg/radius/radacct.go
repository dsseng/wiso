package radius

import "time"

type RadAcct struct {
	RadAcctId           uint64 `gorm:"primary_key;type:bigint;column:radacctid"`
	AcctSessionId       string `gorm:"column:acctsessionid"`       // NOT NULL
	AcctUniqueId        string `gorm:"column:acctuniqueid;unique"` // NOT NULL
	Username            string
	Realm               string
	NASIPAddress        string     `gorm:"column:nasipaddress"` // inet NOT NULL
	NASPortId           string     `gorm:"column:nasportid"`
	NASPortType         string     `gorm:"column:nasporttype"`
	AcctStartTime       *time.Time `gorm:"column:acctstarttime"`
	AcctUpdateTime      *time.Time `gorm:"column:acctupdatetime"`
	AcctStopTime        *time.Time `gorm:"column:acctstoptime"`
	AcctInterval        int64      `gorm:"type:bigint;column:acctinterval"`
	AcctSessionTime     int64      `gorm:"type:bigint;column:acctsessiontime"`
	AcctAuthentic       string     `gorm:"column:acctauthentic"`
	ConnectInfoStart    string     `gorm:"column:connectinfo_start"`
	ConnectInfoStop     string     `gorm:"column:connectinfo_stop"`
	AcctInputOctets     int64      `gorm:"type:bigint;column:acctinputoctets"`
	AcctOutputOctets    int64      `gorm:"type:bigint;column:acctoutputoctets"`
	CalledStationId     string     `gorm:"column:calledstationid"`
	CallingStationId    string     `gorm:"column:callingstationid"`
	AcctTerminateCause  string     `gorm:"column:acctterminatecause"`
	ServiceType         string     `gorm:"column:servicetype"`
	FramedProtocol      string     `gorm:"column:framedprotocol"`
	FramedIPAddress     string     `gorm:"column:framedipaddress"`   // inet
	FramedIPv6Address   string     `gorm:"column:framedipv6address"` // inet
	FramedIPv6Prefix    string     `gorm:"column:framedipv6prefix"`  // inet
	FramedInterfaceId   string     `gorm:"column:framedinterfaceid"`
	DelegatedIPv6Prefix string     `gorm:"column:delegatedipv6prefix"`
	Class               string
}

func (RadAcct) TableName() string {
	return "radacct"
}
