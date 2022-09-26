package runners

import (
	db "back/database"
	"errors"
	"fmt"
	"time"

	gl "github.com/xanzy/go-gitlab"
)

type Runner struct {
	InternalUserId int       `json:"internal_user_id,omitempty" gorm:"internal_user_id"`
	ID             int       `json:"id" gorm:"id"`
	Description    string    `json:"description" gorm:"description"`
	Active         bool      `json:"active" gorm:"active"`
	Paused         bool      `json:"paused" gorm:"paused"`
	Online         bool      `json:"online" gorm:"online"`
	Status         string    `json:"status" gorm:"status"`
	Type           string    `json:"type" gorm:"type"`
	TagList        []string  `json:"tag_list" gorm:"tag_list"`
	RunUntagged    bool      `json:"run_untagged" gorm:"run_untagged"`
	IsShared       bool      `json:"is_shared" gorm:"is_shared"`
	Platform       string    `json:"platform" gorm:"platform"`
	Locked         bool      `json:"locked" gorm:"locked"`
	MaximumTimeout int       `json:"maximum_timeout" gorm:"maximum_timeout"`
	ContactedAt    time.Time `json:"contacted_at" gorm:"contacted_at"`
}

func TranslateGLRunnerDetailsToRunner(rd *gl.RunnerDetails) (Runner, error) {
	if rd == nil {
		return Runner{}, errors.New("Error, invalid gitlab runner details received")
	}
	r := Runner{
		ID:             rd.ID,
		Description:    rd.Description,
		Active:         rd.Active,
		Paused:         rd.Paused,
		Online:         rd.Online,
		Status:         rd.Status,
		Type:           rd.RunnerType,
		TagList:        rd.TagList,
		RunUntagged:    rd.RunUntagged,
		IsShared:       rd.IsShared,
		Platform:       rd.Platform,
		Locked:         rd.Locked,
		MaximumTimeout: rd.MaximumTimeout,
		ContactedAt:    *rd.ContactedAt,
	}

	return r, nil
}

func init() {
	resp := db.DB.AutoMigrate(&Runner{})
	if resp != nil {
		panic(fmt.Sprintf("Error automigrating pipelines table: %s", resp.Error()))
	}
}
