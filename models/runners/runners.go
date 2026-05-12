package runners

import (
	"fmt"
	"time"

	db "github.com/florin-rada/grm-back/database"
)

type Runner struct {
	InternalUserID string    `json:"internal_user_id,omitempty" gorm:"internal_user_id,primaryKey"`
	ID             int64     `json:"id" gorm:"id,primaryKey;autoIncrement:false"`
	Description    string    `json:"description" gorm:"description"`
	Active         bool      `json:"active" gorm:"active"`
	Paused         bool      `json:"paused" gorm:"paused"`
	Online         bool      `json:"online" gorm:"online"`
	Status         string    `json:"status" gorm:"status"`
	RunnerType     string    `json:"runner_type" gorm:"runner_type"`
	TagList        string    `json:"tag_list" gorm:"tag_list"`
	RunUntagged    bool      `json:"run_untagged" gorm:"run_untagged"`
	IsShared       bool      `json:"is_shared" gorm:"is_shared"`
	Platform       string    `json:"platform" gorm:"platform"`
	Locked         bool      `json:"locked" gorm:"locked"`
	MaximumTimeout int64     `json:"maximum_timeout" gorm:"maximum_timeout"`
	ContactedAt    time.Time `json:"contacted_at" gorm:"contacted_at"`
	MaxConcurrent  int64     `json:"max_concurrent" gorm:"max_concurrent"`
	SyncActive     bool      `json:"sync_active" gorm:"sync_active"`
}

func init() {
	// todo: move this to a migrations package
	resp := db.PublicDB.AutoMigrate(&Runner{})
	if resp != nil {
		panic(fmt.Sprintf("Error automigrating runners table: %s", resp.Error()))
	}
}
