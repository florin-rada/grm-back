package jobs

import (
	consterrors "back/pkg/const_errors"
	db "back/pkg/database"
	"back/pkg/models/gitlab"
	"back/pkg/models/synchronized"
	"fmt"
	"time"

	gl "github.com/xanzy/go-gitlab"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type JobsModel struct {
	db     *gorm.DB
	client *gl.Client
}

func NewJobsModel(db *gorm.DB, client *gl.Client) *JobsModel {
	return &JobsModel{
		db:     db,
		client: client,
	}
}

// Job represents a internal row of a job
type Job struct {
	InternalUserId string    `json:"internal_user_id,omitempty" gorm:"internal_user_id,unique,index"`
	ID             int       `json:"id" gorm:"id,unique,index" `
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
	Status          string     `json:"status,omitempty"`
}

type StatisticsSearchArgs struct {
	Names          []string   `json:"names,omitempty"`
	CreatedAtStart *time.Time `json:"created_at_start"`
	CreatedAtEnd   *time.Time `json:"created_at_end"`
	RunnerIDs      []int      `json:"runner_ids,omitempty"`
	ProjectIDs     []int      `json:"project_id,omitempty"`
	Branches       []string   `json:"branches,omitempty"`
	Statuses       []string   `json:"statuses,omitempty"`
	//Stage          string     `json:"stage,omitempty"`
}

func TranslateGLJobToJob(gljob *gl.Job) (*Job, error) {
	if gljob == nil {
		return &Job{}, consterrors.ErrInvalidGitlabJob
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

func (jm JobsModel) GetRunnerJobs(idRunner uint, idUser string, page int, perPage int) ([]Job, error) {
	tr := jm.db.Model(&Job{})
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

func (jm JobsModel) SyncRunnerJobs(runnerID uint) error {
	glJobs, _, err := gitlab.GetRunnerJobs(jm.client, int(runnerID), "", 0, 20)
	if err != nil {
		return err
	}

	jobs := make([]Job, 0, len(glJobs))
	for _, job := range glJobs {
		translated, err := TranslateGLJobToJob(job)
		if err != nil {
			return err
		}
		jobs = append(jobs, *translated)
	}
	resp := jm.db.Clauses(clause.OnConflict{
		UpdateAll: true,
	}).Create(jobs)
	if resp.Error != nil {
		return resp.Error
	}
	return nil
}

func (jm JobsModel) SearchJobsForStatistics(userID string, params StatisticsSearchArgs) ([]Job, error) {
	tr := jm.db.Model(&Job{})
	tr = tr.Where("internal_user_id=?", userID)

	if len(params.Names) == 1 {
		tr = tr.Where("name = ?", params.Names[0])
	} else if len(params.Names) > 1 {
		tr = tr.Where("name = ? ", params.Names)
	}

	if params.CreatedAtStart != nil && params.CreatedAtEnd == nil {
		tr = tr.Where("created_at >= ? ", &params.CreatedAtStart)
	} else if params.CreatedAtStart != nil && params.CreatedAtEnd != nil {
		tr = tr.Where("created_at BETWEEN ? AND ?", &params.CreatedAtStart, &params.CreatedAtEnd)
	} else if params.CreatedAtStart == nil && params.CreatedAtEnd != nil {
		tr = tr.Where("created_at <= ? ", &params.CreatedAtEnd)
	}
	if len(params.RunnerIDs) == 1 {
		tr = tr.Where("runner_id = ? ", params.RunnerIDs[0])
	} else if len(params.RunnerIDs) > 1 {
		tr = tr.Where("runner_id IN ?", params.RunnerIDs)
	}

	if len(params.ProjectIDs) == 1 {
		tr = tr.Where("project_id = ? ", params.ProjectIDs[0])
	} else if len(params.ProjectIDs) > 1 {
		tr = tr.Where("project_id IN ?", params.ProjectIDs)
	}

	if len(params.Branches) == 1 {
		tr = tr.Where("branch = ?", params.Branches[0])
	} else if len(params.Branches) > 1 {
		tr = tr.Where("branch = ?", params.Branches)
	}

	if len(params.Statuses) == 1 {
		tr = tr.Where("status = ?", params.Statuses[0])
	} else if len(params.Statuses) > 1 {
		tr = tr.Where("status = ? ", params.Statuses)
	}

	jobs := []Job{}
	resp := tr.Find(&jobs)
	if resp.Error != nil {
		return []Job{}, resp.Error
	}
	return jobs, nil
}

func (jm JobsModel) SearchRunnerJobs(idRunner uint, userID string, params JobSearchArgs) ([]Job, error) {
	tr := jm.db.Model(&Job{})
	tr = tr.Where("runner_id=?", idRunner)
	tr = tr.Where("internal_user_id=?", userID)

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

func (jm JobsModel) SyncJobsBetweenDates(client *gl.Client, userID int, runnerID int, startDate *time.Time, endDate *time.Time) error {
	if client == nil {
		return consterrors.ErrNoGitClient
	}
	if startDate == nil || endDate == nil {
		return consterrors.ErrNoStartOrEndDate
	}
	truncatedStartDate := startDate.In(time.UTC).Truncate(time.Hour * 24)
	truncatedEndDate := endDate.In(time.UTC).Truncate(time.Hour * 24)

	if truncatedStartDate.After(truncatedEndDate) {
		truncatedStartDate, truncatedEndDate = truncatedEndDate, truncatedStartDate
	}
	sm := synchronized.NewSynchronizedModel(jm.db)
	notSyncedStartDate, notSyncedEndDate, err := sm.GetMinMaxUnsyncedDates(userID, runnerID, &truncatedStartDate, &truncatedEndDate)
	if err != nil {
		return err
	}
	if notSyncedStartDate == nil && notSyncedEndDate == nil {
		return nil
	}

	gljobs, err := gitlab.GetJobsBetween(jm.client, runnerID, notSyncedStartDate, notSyncedEndDate)
	if err != nil {
		return err
	}
	if len(gljobs) == 0 {
		return nil
	}
	jobs := make([]Job, 0, len(gljobs))
	for _, job := range gljobs {
		translated, err := TranslateGLJobToJob(job)
		if err != nil {
			return err
		}
		jobs = append(jobs, *translated)
	}
	resp := jm.db.Clauses(clause.OnConflict{
		UpdateAll: true,
	}).Create(jobs)
	if resp.Error != nil {
		return resp.Error
	}
	err = sm.SaveSyncedDates(userID, runnerID, *notSyncedStartDate, *notSyncedEndDate)
	if err != nil {
		return nil
	}
	return nil
}

func init() {
	resp := db.PublicDB.AutoMigrate(&Job{})
	if resp != nil {
		panic(fmt.Sprintf("Error migrating %s table: %s", "jobs", resp.Error()))
	}
}
