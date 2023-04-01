package gitlab

import (
	"errors"
	"fmt"
	"time"

	gl "github.com/xanzy/go-gitlab"
)

// TODO: replace *gl.CLient with a interface matching the gl.Client functions for unittesting purposes
//var token = "temporary-token"
//var client *gl.Client

var ErrInvalidGitClient = errors.New("invalid gitlab client")

func GetAllRunners(client *gl.Client, page int, perPage int) ([]*gl.Runner, *gl.Response, error) {
	if client == nil {
		return nil, nil, ErrInvalidGitClient
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
		return nil, ErrInvalidGitClient
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

	jobs, resp, err := client.Runners.ListRunnerJobs(runnerID, &lrjo)
	if err != nil {
		return nil, nil, err
	}
	return jobs, resp, nil
}

// GetJobsSince returns a list of all the jobs starting with startDate and to the present day
func GetJobsSince(client *gl.Client, runnerID int, startDate *time.Time) ([]*gl.Job, error) {
	if client == nil {
		return nil, ErrInvalidGitClient
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
		return nil, ErrInvalidGitClient
	}
	jobs := []*gl.Job{}

	if startDate == nil || endDate == nil {
		return []*gl.Job{}, errors.New("invalid start date or end date")
	}

	if endDate.Before(*startDate) {
		return []*gl.Job{}, errors.New("start date is before end date")
	}
	perPage := 2
	// we make sure our start and end date are truncated to only 24 hours
	tmpStartDate := startDate.Truncate(24 * time.Hour)
	startDate = &tmpStartDate
	tmpEndDate := endDate.Truncate(24 * time.Hour)
	endDate = &tmpEndDate
	fmt.Printf("Start Date: %v\nEnd Date: %v", *startDate, *endDate)
	//foundEveryting := false
	// we calculate how many days are between now and our end date
	//numDaysSinceEndDate := int(time.Now().Sub(*endDate) / (24 * time.Hour))
	numDaysSinceEndDate := int(time.Since(*endDate) / (24 * time.Hour))
	//numDaysSinceEndDate := int(time.Since(*endDate) / (24 * time.Hour))
	//numDaysSinceStartDate := int(time.Now().Sub(*startDate) / (24 * time.Hour))
	//var pagesInFirstDay int
	var page int = 1
	var counter = 1
	// we get jobs on first page, if we don't have our "end date" on that page
	// we set page = number of days since today and our end date
	// we get jobs from that page
	// if last item's date is newer that our end date we double our page number by the days
	// until we overshoot our end date
	// After that we keep spliting doing a standard binary search
	// determining our max length
	for {
		page = counter * numDaysSinceEndDate
		// get jobs on page untill we overshoot our end date
		tmpJobs, _, err := client.Runners.ListRunnerJobs(runnerID, &gl.ListRunnerJobsOptions{
			ListOptions: gl.ListOptions{
				Page:    page,
				PerPage: perPage,
			},
		})
		if err != nil {
			return nil, err
		}
		if len(tmpJobs) == 0 {
			numDaysSinceEndDate /= 2
			counter = 1
			continue
		}
		lastJob := tmpJobs[len(tmpJobs)-1]
		fmt.Printf("CreatedAt date for last job in row with id %d is %v", lastJob.ID, lastJob.CreatedAt)
		if lastJob.CreatedAt.Truncate(24*time.Hour).After(*endDate) ||
			lastJob.CreatedAt.Truncate(24*time.Hour).Equal(*endDate) {
			fmt.Printf("Found date after our end date on page %d", page)
			break
		}
		counter++
		//
	}

	minPage := 1
	maxPage := page
	tmpPage := int((maxPage - minPage) / 2)
	for minPage == maxPage {
		tmpJobs, _, err := client.Runners.ListRunnerJobs(runnerID, &gl.ListRunnerJobsOptions{
			ListOptions: gl.ListOptions{
				Page:    tmpPage,
				PerPage: perPage,
			},
		})
		if err != nil {
			return nil, err
		}
		// we know we are on the right side
		if tmpJobs[len(tmpJobs)-1].CreatedAt.Truncate(24 * time.Hour).After(*endDate) {
			minPage = tmpPage
			tmpPage = int((maxPage - minPage) / 2)
			continue
		} else if tmpJobs[len(tmpJobs)-1].CreatedAt.Truncate(24 * time.Hour).Before(*endDate) {
			maxPage = tmpPage
			tmpPage = int((maxPage - minPage) / 2)
			continue
		} else if tmpJobs[len(tmpJobs)-1].CreatedAt.Truncate(24 * time.Hour).Equal(*endDate) {
			// we found our page so we break the loop
			break
		}
	}
	page = tmpPage
	// we might have some more jobs from the same day on the next page so we get the next page and loop throug them adding them to
	// our jobs
	tmpJobs, _, err := client.Runners.ListRunnerJobs(runnerID, &gl.ListRunnerJobsOptions{
		ListOptions: gl.ListOptions{
			Page:    page + 1,
			PerPage: perPage,
		},
	})
	if err != nil {
		return nil, err
	}
	for _, j := range tmpJobs {
		if j.CreatedAt.Truncate(24 * time.Hour).Equal(*endDate) {
			jobs = append(jobs, j)
			continue
		} else if j.CreatedAt.Truncate(24 * time.Hour).After(*endDate) {
			break
		}
	}
	// getting all the jobs until we find a job newer than our start date
	for page >= 1 {
		tmpJobs, _, err := client.Runners.ListRunnerJobs(runnerID, &gl.ListRunnerJobsOptions{
			ListOptions: gl.ListOptions{
				Page:    page,
				PerPage: perPage,
			},
		})
		if err != nil {
			return nil, err
		}
		if tmpJobs[0].CreatedAt.Truncate(24 * time.Hour).Before(*startDate) {
			for i := len(tmpJobs); i >= 0; i-- {
				jobs = append(jobs, tmpJobs[i])
			}
			break
		} else if tmpJobs[0].CreatedAt.Truncate(24 * time.Hour).After(*startDate) {
			jobs = append(tmpJobs, jobs...)
		}
		page--

	}
	return jobs, nil
	// doing the actual binary search part
	/* for pagesInFirstDay < 0 {
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

	return jobs, nil */
}

func GetClient(baseUrl string, token string) (*gl.Client, error) {
	client, err := gl.NewClient(token, gl.WithBaseURL(baseUrl))
	if err != nil {
		return nil, err
	}
	return client, nil
}

func GetAverageDaysPerPage(client *gl.Client, runnerID int) (float64, error) {
	var finished bool
	var count int = 1
	//var pages float64
	jobs := []*gl.Job{}
	for finished && count < 15 {
		tmpJobs, _, err := client.Runners.ListRunnerJobs(runnerID, &gl.ListRunnerJobsOptions{
			ListOptions: gl.ListOptions{
				Page:    count,
				PerPage: 100,
			},
		})
		if err != nil {
			return 0, err
		}
		jobs = append(jobs, tmpJobs...)
		today := time.Now().Truncate(24 * time.Hour)
		lastJob := tmpJobs[len(tmpJobs)-1]
		lastJobCreatedAt := lastJob.CreatedAt.Truncate(24 * time.Hour)
		timeDiff := today.Sub(lastJobCreatedAt).Hours()

		if timeDiff > 48 { // if difference is greater than 48 hours, we loop backwards on the received jobs until we find this

		} else if timeDiff == 48 { // if we have 48 hours, we go back until we get 24 hours difference and than start counting untill we reach today
			var numJobs int
			for i := len(jobs) - 1; i >= 0; i-- {
				tmpTimeDiff := today.Sub(jobs[i].CreatedAt.Truncate(24 * time.Hour)).Hours()
				if tmpTimeDiff < 24 && tmpTimeDiff > 0 {
					numJobs++
				}
				if tmpTimeDiff <= 0 {
					if count == 1 {

					}
				}
			}
		} else if timeDiff < 48 { // we continue to load another page
			continue
		}

	}
	return 0, nil
}

func UpdateRunner(client *gl.Client, id uint, runner *gl.UpdateRunnerDetailsOptions) error {
	if client == nil {
		return ErrInvalidGitClient
	}
	_, _, err := client.Runners.UpdateRunnerDetails(id, runner)
	return err
}

func init() {
	// var err error
	/* client, err = gl.NewClient(token, gl.WithBaseURL("https://gitlab.com/api/v4"))
	if err != nil {
		panic(fmt.Sprintf("Error creating gitlab client: %s", err.Error()))
	} */
	//fmt.Printf("We have client: %+v", client)
}
