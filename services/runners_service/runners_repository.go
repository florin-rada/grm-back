package runnersservice

import (
	consterrors "github.com/florin-rada/grm-back/const_errors"
	"github.com/florin-rada/grm-back/mapper"
	"github.com/florin-rada/grm-back/models/runners"
	gl "gitlab.com/gitlab-org/api/client-go/v2"
	"gorm.io/gorm"
)

type RunnerRepository struct {
	db *gorm.DB
	//client *gl.Client
}

func NewRunnerRepository(db *gorm.DB) *RunnerRepository {
	return &RunnerRepository{db: db}
}

func (rr RunnerRepository) GetRunnersToSync(userID string, maxRunners int) ([]runners.Runner, error) {
	runnersList := []runners.Runner{}
	resp := rr.db.Model(&runners.Runner{}).Where("internal_user_id=?", userID).Limit(maxRunners).Find(&runnersList)
	if resp.Error != nil {
		return []runners.Runner{}, resp.Error
	}
	return runnersList, nil
}

func (rr RunnerRepository) GetRunner(runnerID int64) (runners.Runner, error) {
	r := runners.Runner{}
	resp := rr.db.Find(&r, runnerID)
	if resp.Error != nil {
		return runners.Runner{}, resp.Error
	}
	return r, nil
}

func (rr RunnerRepository) UpdateRunner(r runners.Runner) error {
	resp := rr.db.Save(r)
	return resp.Error
}

// GetUserRunners returns a list of the user's runners extracted from the local database
func (rr RunnerRepository) GetUserRunners(userID string) ([]runners.Runner, error) {
	r := []runners.Runner{}
	resp := rr.db.Where("internal_user_id=?", userID).Find(&r)
	if resp.Error != nil {
		return []runners.Runner{}, resp.Error
	}
	return r, nil
}

// AddRunnerForUser stores a runner's details to the local database
// it accepts the runnerID and based on this, takes the runner details from gitlab
// than stores the relevant runner details to the database
func (rr RunnerRepository) AddRunnerForUser(rd *gl.RunnerDetails, userID string) error {
	if rd == nil {
		return consterrors.ErrInvalidRunnerDetails
	}
	r, err := mapper.TranslateGLRunnerDetailsToRunner(rd)
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

func (rr RunnerRepository) DeleteRunner(runnerID int64, userID string) error {
	resp := rr.db.Delete((&runners.Runner{
		InternalUserID: userID,
		ID:             runnerID,
	}))
	return resp.Error
}
