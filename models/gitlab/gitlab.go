package gitlab

import (
	"fmt"

	gl "github.com/xanzy/go-gitlab"
)

var token = "glpat-SQb5HA2ZVZNC2AUuuyAy"
var client *gl.Client

func GetAllRunners(client *gl.Client, page int, perPage int) ([]*gl.Runner, *gl.Response, error) {
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

func init() {
	var err error
	client, err = gl.NewClient(token, gl.WithBaseURL("https://gitlab.com/api/v4"))
	if err != nil {
		panic(fmt.Sprintf("Error creating gitlab client: %s", err.Error()))
	}
	fmt.Printf("We have client: %+v", client)
}
