package tracking

import "time"

type Event struct {
	ID        int    `gorm:"id,autoincrement,primarykey,unique,index"`
	UUID      string `gorm:"uuid,index"`
	Action    string `gorm:"action"`
	Target    string `gorm:"target"`
	Timestamp time.Time
}
