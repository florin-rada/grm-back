package jobs

import (
	db "back/database"
	"errors"
	"fmt"
	"time"

	gl "github.com/xanzy/go-gitlab"
	"gorm.io/gorm/clause"
)

type Job struct {
	InternalUserId int       `json:"internal_user_id,omitempty" gorm:"internal_user_id,unique"`
	ID             int       `json:"id" gorm:"id,unique" `
	Name           string    `json:"name" gorm:"name"`
	CreatedAt      time.Time `json:"created_at" gorm:"created_at"`
	StartedAt      time.Time `json:"started_at" gorm:"started_at"`
	FinishedAt     time.Time `json:"finished_at" gorm:"finished_at"`
	PipelineID     int       `json:"pipeline_id" gorm:"pipeline_id"`
	ProjectID      int       `json:"project_id" gorm:"project_Id"`
	Branch         string    `json:"branch" gorm:"branch"`
	Duration       float64   `json:"duration" gorm:"duration"`
	QueuedDuration float64   `json:"queued_duration" gorm:"queued_duration"`
	URL            string    `json:"url" gorm:"url"`
	Stage          string    `json:"stage" gorm:"stage"`
}

func (j *Job) Save() error {
	if j.InternalUserId <= 0 {
		return errors.New("Error, User id must be greater than 0")
	}
	if j.ID <= 0 {
		return errors.New("Error, Job ID must be greater than 0")
	}
	resp := db.DB.Clauses(clause.OnConflict{
		UpdateAll: true,
	}).Create(&j)
	if resp.Error != nil {
		return resp.Error
	}
	return nil
}

func TranslateGLJobToJob(gljob *gl.Job) (Job, error) {
	if gljob == nil {
		return Job{}, errors.New("Invalid gitlab job received")
	}
	j := Job{
		ID:             gljob.ID,
		Name:           gljob.Name,
		CreatedAt:      *gljob.CreatedAt,
		StartedAt:      *gljob.StartedAt,
		FinishedAt:     *gljob.FinishedAt,
		PipelineID:     gljob.Pipeline.ID,
		ProjectID:      gljob.Commit.ProjectID,
		Branch:         gljob.Ref,
		Duration:       gljob.Duration,
		QueuedDuration: gljob.QueuedDuration,
		URL:            gljob.WebURL,
		Stage:          gljob.Stage,
	}
	return j, nil
}

func init() {
	resp := db.DB.AutoMigrate(&Job{})
	if resp != nil {
		panic(fmt.Sprintf("Error migrating %s table: %s", "jobs", resp.Error()))
	}
}
