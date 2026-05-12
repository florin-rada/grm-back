package runnersservice

import (
	"fmt"
	"strings"

	const_errors "github.com/florin-rada/grm-back/const_errors"
	"github.com/florin-rada/grm-back/models/runners"
	gl "gitlab.com/gitlab-org/api/client-go/v2"
)

// RunnersService is responsible for handling the business logic related to gitlab runners
// it uses the RunnerRepository to interact with the local database and the gitlab client to interact with gitlab
// the gitlab client is not stored as a field in the service, but is passed as an argument to the methods that need it
// because depending on the user that is making the request, the gitlab client will be different (it will have a different access token)
type RunnersService struct {
	rr *RunnerRepository
}

func NewRunnersService(rr *RunnerRepository) *RunnersService {
	return &RunnersService{rr: rr}
}

func (rs RunnersService) SyncRunnerStatus(client *gl.Client, runnerID int64) error {
	fmt.Printf("Starting updating runner status for %d\n", runnerID)
	r, err := rs.rr.GetRunner(runnerID)
	if err != nil {
		return err
	}
	rd, err := rs.GetRunnerDetails(client, int(runnerID))
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
	return rs.rr.UpdateRunner(r)
}

func (rs RunnersService) GetAllRunners(client *gl.Client, page int64, perPage int64) ([]*gl.Runner, *gl.Response, error) {
	if client == nil {
		return nil, nil, const_errors.ErrInvalidGitClient
	}
	lro := gl.ListRunnersOptions{}
	if page > 0 {
		lro.ListOptions.Page = page
	}
	if perPage > 0 && (perPage < 100 && perPage > 10) {
		lro.ListOptions.PerPage = perPage
	}
	runners, resp, err := client.Runners.ListRunners(&gl.ListRunnersOptions{})
	if err != nil {
		return nil, resp, err
	}
	for idx, r := range runners {
		fmt.Printf("runner: %d: %+v", idx, r)
	}
	return runners, resp, nil
}

func (rs RunnersService) GetRunnerDetails(client *gl.Client, runnerID int) (*gl.RunnerDetails, error) {
	if client == nil {
		return nil, const_errors.ErrInvalidGitClient
	}
	rd, _, err := client.Runners.GetRunnerDetails(runnerID)
	if err != nil {
		return nil, err
	}
	//fmt.Printf("runner details: %+v", rd)
	return rd, nil
}

func (rs RunnersService) UpdateRunnerOnGit(client *gl.Client, r runners.Runner) error {

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

func (rs RunnersService) GetUserRunners(userID string) ([]runners.Runner, error) {
	return rs.rr.GetUserRunners(userID)
}

func (rs RunnersService) DeleteRunner(runnerID int64, userID string) error {
	return rs.rr.DeleteRunner(runnerID, userID)
}

func (rs RunnersService) AddRunnerForUser(client *gl.Client, userID string, runnerID int64) error {
	rd, err := rs.GetRunnerDetails(client, int(runnerID))
	if err != nil {
		return err
	}

	return rs.rr.AddRunnerForUser(rd, userID)
}

func (rs RunnersService) GetRunner(runnerID int64) (runners.Runner, error) {
	return rs.rr.GetRunner(runnerID)
}

func (rs RunnersService) UpdateRunner(r runners.Runner) error {
	return rs.rr.UpdateRunner(r)
}
