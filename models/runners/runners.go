package runners

import (
	consterrors "back/const_errors"
	db "back/database"
	"errors"
	"fmt"
	"strings"
	"time"

	gl "github.com/xanzy/go-gitlab"
	"gorm.io/gorm"
)

type RunnerRepository struct {
	db *gorm.DB
}

type Runner struct {
	InternalUserId string    `json:"internal_user_id,omitempty" gorm:"internal_user_id"`
	ID             int       `json:"id" gorm:"id,primaryKey"`
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
	MaximumTimeout int       `json:"maximum_timeout" gorm:"maximum_timeout"`
	ContactedAt    time.Time `json:"contacted_at" gorm:"contacted_at"`
	MaxConcurrent  int       `json:"max_concurent" gorm:"max_concurrent"`
}

func NewRunnerRepository(db *gorm.DB) *RunnerRepository {
	return &RunnerRepository{db: db}
}

func TranslateGLRunnerDetailsToRunner(rd *gl.RunnerDetails) (*Runner, error) {
	if rd == nil {
		return nil, errors.New("Error, invalid gitlab runner details received")
	}
	r := Runner{
		ID:             rd.ID,
		Description:    rd.Description,
		Active:         rd.Active,
		Paused:         rd.Paused,
		Online:         rd.Online,
		Status:         rd.Status,
		RunnerType:     rd.RunnerType,
		TagList:        strings.Join(rd.TagList, ","),
		RunUntagged:    rd.RunUntagged,
		IsShared:       rd.IsShared,
		Platform:       rd.Platform,
		Locked:         rd.Locked,
		MaximumTimeout: rd.MaximumTimeout,
		ContactedAt:    *rd.ContactedAt,
		MaxConcurrent:  1,
	}

	return &r, nil
}

func (rr RunnerRepository) GetRunner(runnerID uint) (Runner, error) {
	r := Runner{}
	resp := rr.db.Find(&r, runnerID)
	if resp.Error != nil {
		return Runner{}, resp.Error
	}
	return r, nil
}

func (rr RunnerRepository) UpdateRunnerOnGit(r Runner) error {
	return nil
}

func (rr RunnerRepository) UpdateRunner(r Runner) error {
	return nil
}

func (rr RunnerRepository) GetUserRunners(userID string) ([]Runner, error) {
	r := []Runner{}
	resp := rr.db.Where("internal_user_id=?", userID).Find(&r)
	if resp.Error != nil {
		return []Runner{}, resp.Error
	}
	return r, nil
}

func (rr RunnerRepository) AddRunnerForUser(client *gl.Client, userID string, runnerId int) error {
	if client == nil {
		return consterrors.ErrNoGitClient
	}
	if userID == "" {
		return consterrors.ErrNoUserId
	}
	rd, _, err := client.Runners.GetRunnerDetails(runnerId)
	if err != nil {
		return err
	}
	if rd == nil {
		return consterrors.ErrRecordNotFound
	}
	r, err := TranslateGLRunnerDetailsToRunner(rd)
	if err != nil {
		return err
	}
	r.InternalUserId = userID
	resp := rr.db.Save(r)
	if resp.Error != nil {
		return resp.Error
	}
	return nil
}

func init() {
	resp := db.PublicDB.AutoMigrate(&Runner{})
	if resp != nil {
		panic(fmt.Sprintf("Error automigrating runners table: %s", resp.Error()))
	}
}
