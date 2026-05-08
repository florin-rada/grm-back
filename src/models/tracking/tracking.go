package tracking

import (
	"fmt"
	"math"
	"time"

	"github.com/florin-rada/grm-back/src/database"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type EventType string

const (
	Visit        EventType = "visit"
	CallToAction EventType = "call_to_action"
	Leave        EventType = "leave"
	Register     EventType = "register"
)

type Event struct {
	ID        uint      `gorm:"id,autoincrement,primarykey,unique,index"`
	UUID      string    `gorm:"uuid,index"`
	Action    EventType `gorm:"action"`
	Target    string    `gorm:"target"`
	Referrer  string    `gorm:"referrer"`
	Timestamp time.Time `gorm:"timestamp,autoCreateTime"`
}

type Visitor struct {
	UUID    string `gorm:"uuid,primaryKey,index,unique"`
	IP      string `gorm:"ip"`
	Country string `gorm:"country"`
}

type TrackingRepository struct {
	db *gorm.DB
}

func NewTrackingRepository(db *gorm.DB) *TrackingRepository {
	return &TrackingRepository{
		db: db,
	}
}

func (tm TrackingRepository) AddEvent(UUID string, Action EventType, Target string, Referer string) error {
	e := Event{
		UUID:      UUID,
		Action:    Action,
		Target:    Target,
		Timestamp: time.Now(),
		Referrer:  Referer,
	}
	resp := tm.db.Create(&e)
	if resp.Error != nil {
		return resp.Error
	}
	return nil
}

func (tm TrackingRepository) GetConversionRate(startTime *time.Time, endTime *time.Time) (uniqueVisitors int, totalRegistered int, convRate float64, err error) {
	if startTime == nil {
		startTimeTmp := time.Now().In(time.UTC).Add(time.Hour * 24 * -30)
		startTime = &startTimeTmp
	}
	if endTime == nil {
		endTimeTmp := time.Now().In(time.UTC)
		endTime = &endTimeTmp
	}
	tr := tm.db.Model(&Event{})
	tr.Where("timestamp BETWEEN ? and ?", startTime, endTime)
	// Getting unique visits between specified dates
	resp := tr.Select("COUNT(DISTINCT(uuid))").Find(&uniqueVisitors)
	if resp.Error != nil {
		return 0, 0, 0, resp.Error
	}
	tr = tm.db.Model(&Event{})
	tr.Where("timestamp BETWEEN ? and ?", startTime, endTime)
	tr.Where("action = ?", Register)
	resp = tr.Select("COUNT(DISTINCT(uuid))").Find(&totalRegistered)
	if resp.Error != nil {
		return 0, 0, 0, resp.Error
	}
	convRate = math.Round((float64(totalRegistered) / float64(uniqueVisitors)) * 100)
	return uniqueVisitors, totalRegistered, convRate, err
}

func (tm TrackingRepository) GetEvents(UUID string, action EventType, target string, startTime *time.Time, endTime *time.Time, page int, perPage int) ([]Event, error) {
	tr := tm.db.Model(&Event{})

	if UUID != "" {
		tr.Where("uuid LIKE ? ", fmt.Sprintf("%%%s%%", UUID))
	}
	if action != "" {
		tr.Where("action LIKE ?", fmt.Sprintf("%%%s%%", action))
	}

	if target != "" {
		tr.Where("target LIKE ?", fmt.Sprintf("%%%s%%", target))
	}

	if startTime == nil {
		startTimeTmp := time.Now().In(time.UTC).Add(time.Hour * 24 * -30)
		startTime = &startTimeTmp
	}
	if endTime == nil {
		endTimeTmp := time.Now().In(time.UTC)
		endTime = &endTimeTmp
	}
	tr.Where("timestamp BETWEEN ? and ?", startTime, endTime)

	if perPage > 0 {
		if page < 0 {
			page = 0
		}
		tr.Offset(perPage * page).Limit(perPage)
	}

	events := []Event{}
	resp := tm.db.Find(&events)
	if resp.Error != nil {
		return nil, resp.Error
	}
	return events, nil
}

func (tm TrackingRepository) AddVisitor(UUID string, IP string, country string) error {
	v := Visitor{
		UUID:    UUID,
		IP:      IP,
		Country: country,
	}

	resp := tm.db.Clauses(clause.OnConflict{
		DoNothing: true,
	}).Create(v)
	if resp.Error != nil {
		return resp.Error
	}
	return nil
}

func (tm TrackingRepository) GetVisitors(startTime *time.Time, endTime *time.Time, page int, perPage int) ([]Visitor, error) {
	tr := tm.db.Model(&Visitor{})

	if startTime == nil {
		startTimeTmp := time.Now().In(time.UTC).Add(time.Hour * 24 * -30)
		startTime = &startTimeTmp
	}
	if endTime == nil {
		endTimeTmp := time.Now().In(time.UTC)
		endTime = &endTimeTmp
	}
	tr.Where("timestamp BETWEEN ? and ?", startTime, endTime)
	if perPage > 0 {
		if page < 0 {
			page = 0
		}
		tr.Offset(perPage * page).Limit(perPage)
	}

	visitors := []Visitor{}
	resp := tr.Find(&visitors)
	if resp.Error != nil {
		return nil, resp.Error
	}
	return visitors, nil

}

func init() {
	err := database.PublicDB.AutoMigrate(&Event{})
	if err != nil {
		panic(fmt.Sprintf("Error auto migrating Event: %s", err.Error()))
	}
}
