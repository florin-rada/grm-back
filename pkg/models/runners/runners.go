package runners

import (
	consterrors "back/pkg/const_errors"
	db "back/pkg/database"
	"back/pkg/models/gitlab"
	"errors"
	"fmt"
	"strings"
	"time"

	gl "github.com/xanzy/go-gitlab"
	"gorm.io/gorm"
)

type RunnerRepository struct {
	db *gorm.DB
	//client *gl.Client
}

type Runner struct {
	InternalUserID string    `json:"internal_user_id,omitempty" gorm:"internal_user_id,primaryKey"`
	ID             int       `json:"id" gorm:"id,primaryKey;autoIncrement:false"`
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
	SyncActive     bool      `json:"sync_active" gorm:"sync_active"`
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

func (rr RunnerRepository) GetRunnersToSync(userID string, maxRunners int) ([]Runner, error) {
	runners := []Runner{}
	resp := rr.db.Model(&Runner{}).Where("internal_user_id=?", userID).Limit(maxRunners).Find(&runners)
	if resp.Error != nil {
		return []Runner{}, resp.Error
	}
	return runners, nil
}

func (rr RunnerRepository) SyncRunnerStatus(client *gl.Client, runnerID uint) error {
	fmt.Printf("Starting updating runner status for %d\n", runnerID)
	r, err := rr.GetRunner(runnerID)
	if err != nil {
		return err
	}
	rd, err := gitlab.GetRunnerDetails(client, int(runnerID))
	if err != nil {
		return err
	}

	r.Active = rd.Active
	r.Description = rd.Description
	r.Paused = rd.Paused
	r.Online = rd.Online
	r.Status = rd.Status
	r.TagList = strings.Join(rd.TagList, ",")
	r.RunUntagged = rd.RunUntagged
	r.IsShared = rd.IsShared
	r.Locked = rd.Locked
	r.MaximumTimeout = rd.MaximumTimeout
	r.ContactedAt = *rd.ContactedAt
	fmt.Printf("Ending updating runner status for %d\n", runnerID)
	return rr.UpdateRunner(r)
}

func (rr RunnerRepository) GetRunner(runnerID uint) (Runner, error) {
	r := Runner{}
	resp := rr.db.Find(&r, runnerID)
	if resp.Error != nil {
		return Runner{}, resp.Error
	}
	return r, nil
}

func (rr RunnerRepository) UpdateRunnerOnGit(client *gl.Client, r Runner) error {

	tagListArray := strings.Split(r.TagList, ",")
	_, _, err := client.Runners.UpdateRunnerDetails(r.ID, &gl.UpdateRunnerDetailsOptions{
		Description:    &r.Description,
		Paused:         &r.Paused,
		TagList:        &tagListArray,
		RunUntagged:    &r.RunUntagged,
		MaximumTimeout: &r.MaximumTimeout,
		Locked:         &r.Locked,
	})
	return err
}

func (rr RunnerRepository) UpdateRunner(r Runner) error {
	resp := rr.db.Save(r)
	return resp.Error
}

// GetUserRunners returns a list of the user's runners extracted from the local database
func (rr RunnerRepository) GetUserRunners(userID string) ([]Runner, error) {
	r := []Runner{}
	resp := rr.db.Where("internal_user_id=?", userID).Find(&r)
	if resp.Error != nil {
		return []Runner{}, resp.Error
	}
	return r, nil
}

// AddRunnerForUser stores a runner's details to the local database
// it accepts the runnerID and based on this, takes the runner details from gitlab
// than stores the relevant runner details to the database
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
	r.InternalUserID = userID
	resp := rr.db.Save(r)
	if resp.Error != nil {
		return resp.Error
	}
	return nil
}

func (rr RunnerRepository) DeleteRunner(runnerID uint, userID string) error {
	resp := rr.db.Delete((&Runner{
		InternalUserID: userID,
		ID:             int(runnerID),
	}))
	return resp.Error
}

func init() {
	resp := db.PublicDB.AutoMigrate(&Runner{})
	if resp != nil {
		panic(fmt.Sprintf("Error automigrating runners table: %s", resp.Error()))
	}
}
