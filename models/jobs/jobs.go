package jobs

import (
	db "back/database"
	"errors"
	"fmt"
	"time"

	gl "github.com/xanzy/go-gitlab"
	"gorm.io/gorm/clause"
)

// Job represents a internal row of a job
type Job struct {
	InternalUserId int       `json:"internal_user_id,omitempty" gorm:"internal_user_id,unique"`
	ID             int       `json:"id" gorm:"id,unique" `
	Name           string    `json:"name" gorm:"name"`
	CreatedAt      time.Time `json:"created_at" gorm:"created_at"`
	StartedAt      time.Time `json:"started_at" gorm:"started_at"`
	FinishedAt     time.Time `json:"finished_at" gorm:"finished_at"`
	PipelineID     int       `json:"pipeline_id" gorm:"pipeline_id"`
	ProjectID      int       `json:"project_id" gorm:"project_Id"`
	RunnerID       int       `json:"runner_id" gorm:"runner_id"`
	Branch         string    `json:"branch" gorm:"branch"`
	Duration       float64   `json:"duration" gorm:"duration"`
	QueuedDuration float64   `json:"queued_duration" gorm:"queued_duration"`
	URL            string    `json:"url" gorm:"url"`
	Stage          string    `json:"stage" gorm:"stage"`
}

// JobSearchArgs is a struct used for filtering jobs by job related criteria
// It will not include the user or the runner
type JobSearchArgs struct {
	Name            string     `json:"name,omitempty"`
	CreatedAtStart  *time.Time `json:"created_at_start,omitempty"`
	CreatedAtEnd    *time.Time `json:"created_at_end,omitempty"`
	StartedAtStart  *time.Time `json:"started_at_start,omitempty"`
	StartedAtEnd    *time.Time `json:"started_at_end,omitempty"`
	FinishedAtStart *time.Time `json:"finished_at_start,omitempty"`
	FinishedAtEnd   *time.Time `json:"finished_at_end,omitempty"`
	PipelineID      int        `json:"pipeline_id,omitempty"`
	ProjectID       int        `json:"project_id,omitempty"`
	Branch          string     `json:"branch,omitempty"`
	Stage           string     `json:"stage,omitempty"`
	Page            int        `json:"page,omitempty"`
	PerPage         int        `json:"per_page,omitempty"`
}

type JobsQueueItem struct {
	Task func()
}

func (jqi JobsQueueItem) ExecuteTask() {
	jqi.Task()
}

func (j *Job) Save() error {
	if j.InternalUserId <= 0 {
		return errors.New("Error, User id must be greater than 0")
	}
	if j.ID <= 0 {
		return errors.New("Error, Job ID must be greater than 0")
	}
	resp := db.PublicDB.Clauses(clause.OnConflict{
		UpdateAll: true,
	}).Create(&j)
	if resp.Error != nil {
		return resp.Error
	}
	return nil
}

func TranslateGLJobToJob(gljob *gl.Job) (*Job, error) {
	if gljob == nil {
		return &Job{}, errors.New("Invalid gitlab job received")
	}
	j := Job{
		ID:             gljob.ID,
		Name:           gljob.Name,
		CreatedAt:      *gljob.CreatedAt,
		StartedAt:      *gljob.StartedAt,
		FinishedAt:     *gljob.FinishedAt,
		PipelineID:     gljob.Pipeline.ID,
		ProjectID:      gljob.Commit.ProjectID,
		RunnerID:       gljob.Runner.ID,
		Branch:         gljob.Ref,
		Duration:       gljob.Duration,
		QueuedDuration: gljob.QueuedDuration,
		URL:            gljob.WebURL,
		Stage:          gljob.Stage,
	}
	return &j, nil
}

func GetRunnerJobs(idRunner uint, idUser uint, page int, perPage int) ([]Job, error) {
	tr := db.PublicDB.Model(&Job{})
	tr = tr.Where("runner_id=?", idRunner)
	tr = tr.Where("internal_user_id=?", idUser)
	if perPage < 0 {
		perPage = 25
	}
	if page > 0 {
		tr = tr.Offset(page * perPage).Limit(perPage)
	}
	jobs := []Job{}
	resp := tr.Find(&jobs)
	if resp.Error != nil {
		return []Job{}, resp.Error
	}
	return jobs, nil
}

func SearchRunnerJobs(idRunner uint, idUser uint, params JobSearchArgs) ([]Job, error) {
	tr := db.PublicDB.Model(&Job{})
	tr = tr.Where("runner_id=?", idRunner)
	tr = tr.Where("internal_user_id=?", idUser)

	if params.Name != "" {
		tr = tr.Where("name Like ?", fmt.Sprintf("%%%s%%", params.Name))
	}
	if params.ProjectID != 0 && params.ProjectID > 0 {
		tr = tr.Where("project_id = ? ", params.ProjectID)
	}
	if params.PipelineID != 0 && params.PipelineID > 0 {
		tr = tr.Where("pipeline_id = ?", params.PipelineID)
	}
	if params.CreatedAtStart != nil && params.CreatedAtEnd == nil {
		tr = tr.Where("created_at >= ? ", &params.CreatedAtStart)
	} else if params.CreatedAtStart != nil && params.CreatedAtEnd != nil {
		tr = tr.Where("created_at BETWEEN ? AND ?", &params.CreatedAtStart, &params.CreatedAtEnd)
	} else if params.CreatedAtStart == nil && params.CreatedAtEnd != nil {
		tr = tr.Where("created_at <= ? ", &params.CreatedAtEnd)
	}

	if params.StartedAtStart != nil && params.StartedAtEnd == nil {
		tr = tr.Where("started_at >= ? ", &params.StartedAtStart)
	} else if params.StartedAtStart != nil && params.StartedAtEnd != nil {
		tr = tr.Where("started_at BETWEEN ? AND ?", &params.StartedAtStart, &params.StartedAtEnd)
	} else if params.StartedAtStart == nil && params.StartedAtEnd != nil {
		tr = tr.Where("started_at <= ? ", &params.StartedAtEnd)
	}

	if params.FinishedAtStart != nil && params.FinishedAtEnd == nil {
		tr = tr.Where("finished_at >= ? ", &params.FinishedAtStart)
	} else if params.FinishedAtStart != nil && params.FinishedAtEnd != nil {
		tr = tr.Where("finished_at BETWEEN ? AND ?", &params.FinishedAtStart, &params.FinishedAtEnd)
	} else if params.FinishedAtStart == nil && params.FinishedAtEnd != nil {
		tr = tr.Where("finished_at <= ? ", &params.FinishedAtEnd)
	}
	if params.Stage != "" {
		tr = tr.Where("stage LIKE ? ", fmt.Sprintf("%%%s%%", params.Stage))
	}
	if params.Branch != "" {
		tr = tr.Where("branch LIKE ? ", fmt.Sprintf("%%%s%%", params.Branch))
	}

	if params.PerPage < 0 {
		params.PerPage = 25
	}
	if params.Page > 0 {
		tr = tr.Offset(params.Page * params.PerPage).Limit(params.PerPage)
	}

	jobs := []Job{}
	resp := tr.Find(&jobs)
	if resp.Error != nil {
		return []Job{}, resp.Error
	}
	return jobs, nil
}

func init() {
	resp := db.PublicDB.AutoMigrate(&Job{})
	if resp != nil {
		panic(fmt.Sprintf("Error migrating %s table: %s", "jobs", resp.Error()))
	}
}
