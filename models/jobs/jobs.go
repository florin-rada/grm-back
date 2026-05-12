package jobs

import (
	"fmt"
	"time"

	db "github.com/florin-rada/grm-back/database"
)

// Job represents a internal row of a job
type Job struct {
	InternalUserID string    `json:"internal_user_id,omitempty" gorm:"internal_user_id,unique,index"`
	ID             int64     `json:"id" gorm:"id,unique,index" `
	Name           string    `json:"name" gorm:"name"`
	CreatedAt      time.Time `json:"created_at" gorm:"created_at"`
	StartedAt      time.Time `json:"started_at" gorm:"started_at"`
	FinishedAt     time.Time `json:"finished_at" gorm:"finished_at"`
	PipelineID     int64     `json:"pipeline_id" gorm:"pipeline_id"`
	ProjectID      int64     `json:"project_id" gorm:"project_id"`
	RunnerID       int64     `json:"runner_id" gorm:"runner_id"`
	Branch         string    `json:"branch" gorm:"branch"`
	Duration       float64   `json:"duration" gorm:"duration"`
	QueuedDuration float64   `json:"queued_duration" gorm:"queued_duration"`
	URL            string    `json:"url" gorm:"url"`
	Stage          string    `json:"stage" gorm:"stage"`
	Status         string    `json:"status" gorm:"status"`
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
	PipelineID      int64      `json:"pipeline_id,omitempty"`
	ProjectID       int64      `json:"project_id,omitempty"`
	Branch          string     `json:"branch,omitempty"`
	Stage           string     `json:"stage,omitempty"`
	Page            int64      `json:"page,omitempty"`
	PerPage         int64      `json:"per_page,omitempty"`
	Status          string     `json:"status,omitempty"`
}

type StatisticsSearchArgs struct {
	Names          []string   `json:"names,omitempty"`
	CreatedAtStart *time.Time `json:"created_at_start"`
	CreatedAtEnd   *time.Time `json:"created_at_end"`
	RunnerIDs      []int64    `json:"runner_ids,omitempty"`
	ProjectIDs     []int64    `json:"project_ids,omitempty"`
	Branches       []string   `json:"branches,omitempty"`
	Statuses       []string   `json:"statuses,omitempty"`
	//Stage          string     `json:"stage,omitempty"`
}

func init() {
	// this needs to be moved to a migrations step
	resp := db.PublicDB.AutoMigrate(&Job{})
	if resp != nil {
		panic(fmt.Sprintf("Error migrating %s table: %s", "jobs", resp.Error()))
	}
}
