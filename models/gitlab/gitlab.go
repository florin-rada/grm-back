package gitlab

import (
	"errors"
	"fmt"
	"time"

	gl "github.com/xanzy/go-gitlab"
)

var token = "glpat-SQb5HA2ZVZNC2AUuuyAy"
var client *gl.Client

func GetAllRunners(client *gl.Client, page int, perPage int) ([]*gl.Runner, *gl.Response, error) {
	if client == nil {
		return nil, nil, errors.New("Error, invalid gitlab client")
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

func GetRunnerDetails(client *gl.Client, runnerID int) (*gl.RunnerDetails, error) {
	if client == nil {
		return nil, errors.New("Error, invalid gitlab client")
	}
	rd, _, err := client.Runners.GetRunnerDetails(runnerID)
	if err != nil {
		return nil, err
	}
	fmt.Printf("runner details: %+v", rd)
	return rd, nil
}

func GetRunnerJobs(client *gl.Client, runnerID int, status string, page int, perPage int) ([]*gl.Job, *gl.Response, error) {
	lrjo := gl.ListRunnerJobsOptions{}
	if status != "" {
		lrjo.Status = &status
	}
	if page > 0 {
		lrjo.ListOptions.Page = page
	}
	if perPage > 0 && (perPage < 100 && perPage > 10) {
		lrjo.ListOptions.PerPage = perPage
	}

	jobs, resp, err := client.Runners.ListRunnerJobs(runnerID, &gl.ListRunnerJobsOptions{})
	if err != nil {
		return nil, nil, err
	}
	return jobs, resp, nil
}

// GetJobsSince returns a list of all the jobs starting with startDate and to the present day
func GetJobsSince(client *gl.Client, runnerID int, startDate *time.Time) ([]*gl.Job, error) {
	if client == nil {
		return nil, errors.New("Error, invalid gitlab client")
	}
	jobs := []*gl.Job{}
	page := 0
	foundBeforeDate := false
	for !foundBeforeDate {
		tmpJobs, _, err := client.Runners.ListRunnerJobs(runnerID, &gl.ListRunnerJobsOptions{
			ListOptions: gl.ListOptions{
				Page:    page,
				PerPage: 100,
			},
		})
		if err != nil {
			return nil, err
		}
		for _, tmpJob := range tmpJobs {
			if !startDate.Before(*tmpJob.CreatedAt) {
				foundBeforeDate = true
				break
			}
			jobs = append(jobs, tmpJob)
		}
	}
	return jobs, nil
}

func GetJobsBetween(client *gl.Client, runnerID int, startDate *time.Time, endDate *time.Time) ([]*gl.Job, error) {
	if client == nil {
		return nil, errors.New("Error, invalid gitlab client")
	}
	jobs := []*gl.Job{}

	if startDate == nil || endDate == nil {
		return []*gl.Job{}, errors.New("Error, invalid start date or end date")
	}

	if endDate.Before(*startDate) {
		return []*gl.Job{}, errors.New("Error, start date is before end date")
	}
	//foundEveryting := false
	//numDaysSinceEndDate := int(time.Now().Sub(*endDate) / (24 * time.Hour))
	numDaysSinceEndDate := int(time.Since(*endDate) / (24 * time.Hour))
	//numDaysSinceStartDate := int(time.Now().Sub(*startDate) / (24 * time.Hour))
	var pagesInFirstDay int
	var page int = 1
	for pagesInFirstDay < 0 {
		tmpJobs, _, err := client.Runners.ListRunnerJobs(runnerID, &gl.ListRunnerJobsOptions{
			ListOptions: gl.ListOptions{
				Page:    page,
				PerPage: 100,
			},
		})
		if err != nil {
			return nil, err
		}
		lastJob := tmpJobs[len(tmpJobs)-1]
		daysInPage := int(time.Since(*lastJob.CreatedAt) / (24 * time.Hour))
		if daysInPage > 0 {
			pagesInFirstDay = page
			break
		}
		page++
	}
	newPage := numDaysSinceEndDate * pagesInFirstDay
	var foundBeforeDate bool = false
	//var foundEndDate bool = false
	for foundBeforeDate {
		if newPage < 0 {
			break
		}
		tmpJobs, _, err := client.Runners.ListRunnerJobs(runnerID, &gl.ListRunnerJobsOptions{
			ListOptions: gl.ListOptions{
				Page:    newPage,
				PerPage: 100,
			},
		})
		if err != nil {
			return nil, err
		}
		if len(tmpJobs) == 0 {
			newPage--
			continue
		}
		tmpJob := tmpJobs[0]
		if !endDate.Before(*tmpJob.CreatedAt) {
			newPage--
			continue
		}
		for _, tmpJob := range tmpJobs {
			if startDate.Before(*tmpJob.CreatedAt) {
				foundBeforeDate = true
				break
			}
			if endDate.Before(*tmpJob.CreatedAt) {
				jobs = append(jobs, tmpJob)
			}
		}

	}

	return jobs, nil

}

func init() {
	var err error
	client, err = gl.NewClient(token, gl.WithBaseURL("https://gitlab.com/api/v4"))
	if err != nil {
		panic(fmt.Sprintf("Error creating gitlab client: %s", err.Error()))
	}
	fmt.Printf("We have client: %+v", client)
}
