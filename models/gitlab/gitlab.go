package gitlab

import (
	consterrors "github.com/florin-rada/grm-back/const_errors"
	gl "gitlab.com/gitlab-org/api/client-go/v2"
)

func GetRunnerDetails(client *gl.Client, runnerID int) (*gl.RunnerDetails, error) {
	if client == nil {
		return nil, consterrors.ErrInvalidGitClient
	}
	rd, _, err := client.Runners.GetRunnerDetails(runnerID)
	if err != nil {
		return nil, err
	}
	//fmt.Printf("runner details: %+v", rd)
	return rd, nil
}

func GetClient(baseUrl string, token string) (*gl.Client, error) {
	client, err := gl.NewClient(token, gl.WithBaseURL(baseUrl))
	if err != nil {
		return nil, err
	}
	return client, nil
}

func init() {
	// var err error
	/* client, err = gl.NewClient(token, gl.WithBaseURL("https://gitlab.com/api/v4"))
	if err != nil {
		panic(fmt.Sprintf("Error creating gitlab client: %s", err.Error()))
	} */
	//fmt.Printf("We have client: %+v", client)
}
