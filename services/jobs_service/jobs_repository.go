package jobsservice

import (
	"fmt"

	jobs_model "github.com/florin-rada/grm-back/models/jobs"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type JobsRepository struct {
	db *gorm.DB
}

func NewJobsRepository(db *gorm.DB) *JobsRepository {
	return &JobsRepository{db: db}
}

func (jr *JobsRepository) SearchJobsForStatistics(userID string, params jobs_model.StatisticsSearchArgs) ([]jobs_model.Job, error) {
	tr := jr.db.Model(&jobs_model.Job{})
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

	jobs := []jobs_model.Job{}
	resp := tr.Find(&jobs)
	if resp.Error != nil {
		return []jobs_model.Job{}, resp.Error
	}
	return jobs, nil
}

func (jr JobsRepository) SearchRunnerJobs(idRunner int64, userID string, params jobs_model.JobSearchArgs) ([]jobs_model.Job, error) {
	tr := jr.db.Model(&jobs_model.Job{})
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
		tr = tr.Offset(int(params.Page * params.PerPage)).Limit(int(params.PerPage))
	}

	jobs := []jobs_model.Job{}
	resp := tr.Find(&jobs)
	if resp.Error != nil {
		return []jobs_model.Job{}, resp.Error
	}
	return jobs, nil
}

func (jr JobsRepository) GetRunnerJobs(idRunner uint, idUser string, page int, perPage int, order string) ([]jobs_model.Job, error) {
	tr := jr.db.Model(&jobs_model.Job{})
	tr = tr.Where("runner_id=?", idRunner)
	tr = tr.Where("internal_user_id=?", idUser)
	if perPage < 0 {
		perPage = 25
	}
	if page > 0 {
		tr = tr.Offset(page * perPage).Limit(perPage)
	}
	if order != "" {
		if order == "asc" {
			tr = tr.Order("id asc")
		} else if order == "desc" {
			tr = tr.Order("id desc")
		}
	}
	jobs := []jobs_model.Job{}
	resp := tr.Find(&jobs)
	if resp.Error != nil {
		return []jobs_model.Job{}, resp.Error
	}
	return jobs, nil
}

func (jr JobsRepository) UpsertJobs(jobs []jobs_model.Job) error {
	resp := jr.db.Clauses(clause.OnConflict{
		UpdateAll: true,
	}).Create(jobs)
	if resp.Error != nil {
		return resp.Error
	}
	return nil
}
