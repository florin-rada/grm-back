package synchronized

import (
	"errors"
	"fmt"
	"time"

	"github.com/florin-rada/grm-back/database"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type Synchronized struct {
	UserID   string    `gorm:"user_id,index:idx_user_runner,uniqueIndex:single_sync_date"`
	RunnerID int64     `gorm:"runner_id,index:idx_user_runner,uniqueIndex:single_sync_date"`
	Date     time.Time `gorm:"date,uniqueIndex:single_sync_date"`
}

type SynchronizedModel struct {
	db *gorm.DB
}

func NewSynchronizedModel(db *gorm.DB) *SynchronizedModel {
	return &SynchronizedModel{db: db}
}

func (sm *SynchronizedModel) GetMinMaxUnsyncedDates(userID string, runnerID int64, startDate *time.Time, endDate *time.Time) (*time.Time, *time.Time, error) {
	if startDate == nil || endDate == nil {
		return nil, nil, errors.New("no start date or end date")
	}

	syncedDates := []Synchronized{}
	resp := sm.db.Model(&Synchronized{}).Where("user_id = ? and runner_id = ? and (synced_date BETWEEN ? and ?)", userID, runnerID, startDate.Format("2006-01-02"), endDate.Format("2006-01-02"), runnerID).Order("date desc").Find(&syncedDates)
	if resp.Error != nil {
		return nil, nil, resp.Error
	}
	if len(syncedDates) == 0 {
		return startDate, endDate, nil
	}

	notSyncedDates := []time.Time{}
	currentDate := *startDate
	for endDate.Before(currentDate) {
		for _, date := range syncedDates {
			tmpDate := time.Time(date.Date)
			found := false
			if tmpDate.Year() == currentDate.Year() ||
				tmpDate.Month() == currentDate.Month() ||
				tmpDate.Day() == currentDate.Day() {
				found = true
			}

			if !found {
				notSyncedDates = append(notSyncedDates, currentDate)
			}
			currentDate = currentDate.Add(time.Hour * -24)
		}
	}
	if len(notSyncedDates) == 0 {
		return nil, nil, nil
	}

	return &notSyncedDates[0], &notSyncedDates[len(notSyncedDates)-1], nil
}

func (sm *SynchronizedModel) SaveSyncedDates(userID string, runnerID int64, startDate time.Time, endDate time.Time) error {
	numDays := endDate.Sub(startDate).Hours() / 24
	syncDates := make([]Synchronized, 0, int(numDays))
	currentDay := startDate
	for currentDay.Before(endDate) {
		syncDate := Synchronized{
			UserID:   userID,
			RunnerID: runnerID,
			Date:     currentDay,
		}
		syncDates = append(syncDates, syncDate)
		currentDay = currentDay.Add(time.Hour * 24)
	}
	resp := sm.db.Clauses(clause.OnConflict{
		UpdateAll: true,
	}).Create(syncDates)
	if resp.Error != nil {
		return resp.Error
	}
	return nil
}

func init() {
	err := database.PublicDB.AutoMigrate(&Synchronized{})
	if err != nil {
		panic(fmt.Errorf("Error migrating Synchronized db: %s", err.Error()))
	}
}
