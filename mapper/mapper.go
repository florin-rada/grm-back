package mapper

import (
	"errors"
	"strings"

	consterrors "github.com/florin-rada/grm-back/const_errors"
	jobs_model "github.com/florin-rada/grm-back/models/jobs"
	"github.com/florin-rada/grm-back/models/runners"
	gl "gitlab.com/gitlab-org/api/client-go/v2"
)

func TranslateGLRunnerDetailsToRunner(rd *gl.RunnerDetails) (*runners.Runner, error) {
	if rd == nil {
		return nil, errors.New("Error, invalid gitlab runner details received")
	}
	r := runners.Runner{
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

func TranslateGLJobToJob(userID string, gljob *gl.Job) (*jobs_model.Job, error) {
	if gljob == nil {
		return &jobs_model.Job{}, consterrors.ErrInvalidGitlabJob
	}
	j := jobs_model.Job{
		InternalUserID: userID,
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
		Status:         gljob.Status,
	}
	return &j, nil
}
