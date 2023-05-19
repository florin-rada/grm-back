package gitlab

import (
	"errors"
	"fmt"
	"sort"
	"strconv"
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

// GetJobsMaxPageForDate will try and determine what is the max page for
// our binary search. It will iterate the number of days times a counter
// in it either finds a empty page or a page with jobs older than our
// target date
func GetJobsMaxPageForDate(client *gl.Client, runnerID int, endDate time.Time) (int, error) {
	fmt.Printf("\nGetJobsMaxPageForDate: Starting for RunnerID: %d endDate: %v", runnerID, endDate)
	defer fmt.Printf("\nGetJobsMaxPageForDate: Ended for RunnerID: %d, endDate: %v", runnerID, endDate)
	pageCounter := int(time.Since(endDate) / (24 * time.Hour))
	perPage := 100
	sortDirection := "desc"
	orderBy := "id"
	endDate = endDate.In(time.UTC).Truncate(time.Hour * 24)
	for {
		tmpJobs, _, err := client.Runners.ListRunnerJobs(runnerID, &gl.ListRunnerJobsOptions{
			ListOptions: gl.ListOptions{
				Page:    pageCounter,
				PerPage: perPage,
			},
			OrderBy: &orderBy,
			Sort:    &sortDirection,
		})
		if err != nil {
			return 0, err
		}
		if len(tmpJobs) == 0 || tmpJobs[len(tmpJobs)-1].CreatedAt.In(time.UTC).Truncate(time.Hour*24).Before(endDate) {
			return pageCounter, nil
		}
		pageCounter = pageCounter * 2
	}
}

// GetJobsTartegPage performs a binary search to find the page that has jobs
// with CreatedAt between our start and end dates
// It will return a page number that contains at least a job between
// our target start and end dates
// GetJobsTartegPage performs a binary search to find the page that has jobs
// with CreatedAt between our start and end dates
// It will return a page number that contains at least a job between
// our target start and end dates
func GetJobsTargetPage(client *gl.Client, runnerID, maxPage int, startDate time.Time, endDate time.Time) (int, error) {
	fmt.Printf("GetJobsTargetPage started for RunnerID: %d, startDate: %v, endDate: %v", runnerID, startDate, endDate)
	defer fmt.Printf("GetJobsTargetPage finished for RunnerID: %d, startDate: %v, endDate: %v", runnerID, startDate, endDate)
	minPage := 1
	orderBy := "id"
	sortDirection := "desc"
	prevMaxPage := minPage
	prevMinPage := maxPage
	for minPage <= maxPage {
		currentPage := int((maxPage + minPage) / 2)
		tmpJobs, _, err := client.Runners.ListRunnerJobs(runnerID, &gl.ListRunnerJobsOptions{
			ListOptions: gl.ListOptions{
				Page:    currentPage,
				PerPage: 100,
			},
			OrderBy: &orderBy,
			Sort:    &sortDirection,
		})
		if err != nil {
			return 0, err
		}

		if len(tmpJobs) == 0 {
			if maxPage == prevMaxPage && minPage == prevMinPage {
				break
			}
			prevMaxPage = maxPage
			maxPage = (currentPage - 1)
			continue
		}
		if HaveJobBetweenDates(tmpJobs, startDate, endDate) {
			return currentPage, nil
		}
		middleJobCreatedAt := tmpJobs[int(len(tmpJobs)/2)].CreatedAt.In(time.UTC).Truncate(time.Hour * 24)
		if middleJobCreatedAt.After(endDate) {
			if maxPage == prevMaxPage && minPage == prevMinPage {
				break
			}
			prevMinPage = minPage
			minPage = currentPage + 1

			continue
		}
		if middleJobCreatedAt.Before(endDate) {
			if maxPage == prevMaxPage && minPage == prevMinPage {
				break
			}
			prevMaxPage = maxPage
			maxPage = (currentPage - 1)
			continue
		}
	}
	return maxPage, nil
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
	// due to moving nature of the results returned from the api
	// sometimes we the same job can apear in two results
	// to eliminate this we store the jobs in a map with the ID as the key
	// and the ids in a slice
	// this way at the end we can sort the ID slice and than add them to our slice result
	jobsMap := make(map[int]*gl.Job)
	jobIds := []int{}
	if startDate == nil || endDate == nil {
		return []*gl.Job{}, errors.New("invalid start date or end date")
	}

	/* if endDate.Before(*startDate) {
		return []*gl.Job{}, errors.New("start date is before end date")
	} */
	perPage := 100
	// we make sure our start and end date are truncated to only 24 hours
	tmpStartDate := startDate.Truncate(24 * time.Hour)
	startDate = &tmpStartDate
	tmpEndDate := endDate.Truncate(24 * time.Hour)
	endDate = &tmpEndDate
	fmt.Printf("Start Date: %v\nEnd Date: %v", *startDate, *endDate)
	//foundEveryting := false
	// we calculate how many days are between now and our end date
	//numDaysSinceEndDate := int(time.Since(*endDate) / (24 * time.Hour))
	//var pagesInFirstDay int
	var page int = 1
	//var counter = 1
	// we first get the first page
	// if we receive less than 100 jobs than we can just iterate over
	// what we received and return the jobs in our range
	sortDir := "desc"
	orderBy := "id"
	tmpJobs, resp, err := client.Runners.ListRunnerJobs(runnerID, &gl.ListRunnerJobsOptions{
		ListOptions: gl.ListOptions{
			Page:    page,
			PerPage: perPage,
		},
		OrderBy: &orderBy,
		Sort:    &sortDir,
	})
	if err != nil {
		return nil, err
	}
	if len(tmpJobs) == 0 {
		return []*gl.Job{}, nil
	}
	firstJobCreatedAt := tmpJobs[0].CreatedAt.Truncate(time.Hour * 24)
	if endDate.After(firstJobCreatedAt) {
		return []*gl.Job{}, nil
	}
	if len(tmpJobs) < perPage {
		for _, job := range tmpJobs {
			createdAt := job.CreatedAt.Truncate(time.Hour * 24)
			fmt.Printf("createdAt truncated: %v\n", createdAt)
			if (createdAt.After(*startDate) || createdAt.Equal(*startDate)) && (createdAt.Before(*endDate) || createdAt.Equal(*endDate)) {
				jobs = append([]*gl.Job{job}, jobs...)
			}
		}
		return jobs, nil
	}
	// If we don't have less than 100 jobs on the first page
	// We check if we have a x-total-page key in the header of the response
	// not all apis will have this
	// If we have this we know what our max pages are so we know where to start searching
	maxPageStr := resp.Header.Get("x-total-pages")
	var maxPage int
	if maxPageStr != "" {
		tmp, err := strconv.ParseInt(maxPageStr, 10, 64)
		if err != nil {
			fmt.Printf("Warning, no x-total-pages header received, skipping\n")
		} else {
			maxPage = int(tmp)
		}
	}
	fmt.Printf("Date of last job in tmpJobs: %v", tmpJobs[len(tmpJobs)-1].CreatedAt.Truncate(time.Hour*24))
	fmt.Printf("Num pages as received from x-total-page: %d", maxPage)
	// we know we don't have anything on the first page so we try to find our desired pages
	// try to find max limit by multiplying the number of days between today and end date
	// with a counter until we either receive a empty list or a date older than end date
	//minPage := 1
	if maxPage == 0 {
		maxPage, err = GetJobsMaxPageForDate(client, runnerID, *endDate)
		if err != nil {
			return nil, err
		}
	}
	fmt.Printf("Starting binary search for our targeted dates")

	maxPage, err = GetJobsTargetPage(client, runnerID, maxPage, *startDate, *endDate)
	if err != nil {
		return nil, err
	}

	fmt.Printf("Finished binary search for targeted dates, maxPage is : %d", maxPage)
	// going backward maxPage-- and adding all jobs
	fmt.Printf("starting get jobs newer than end date")
	pageCounter := maxPage
prevPages:
	for pageCounter >= 0 {
		tmpJobs, _, err := client.Runners.ListRunnerJobs(runnerID, &gl.ListRunnerJobsOptions{
			ListOptions: gl.ListOptions{
				Page:    pageCounter,
				PerPage: perPage,
			},
			OrderBy: &orderBy,
			Sort:    &sortDir,
		})
		if err != nil {
			return nil, err
		}
		foundNewer := false
		for _, job := range tmpJobs {
			if job.CreatedAt.Truncate(time.Hour * 24).After(*endDate) {
				foundNewer = true
				continue
			}
			createdAt := job.CreatedAt.Truncate(time.Hour * 24)
			if (createdAt.After(*startDate) || createdAt.Equal(*startDate)) && (createdAt.Before(*endDate) || createdAt.Equal(*endDate)) {
				//jobs = append([]*gl.Job{job}, jobs...)
				if _, ok := jobsMap[job.ID]; !ok {
					jobsMap[job.ID] = job
					jobIds = append(jobIds, job.ID)
				}
			}
		}
		if foundNewer {
			break prevPages
		}
		pageCounter--
	}
	fmt.Printf("Finished getting jobs newer than end date")
	fmt.Printf("Starting getting rest of jobs until we pass end date")
	// going forward maxPage++ and adding all jobs
	pageCounter = maxPage
nextPages:
	for {
		tmpJobs, _, err := client.Runners.ListRunnerJobs(runnerID, &gl.ListRunnerJobsOptions{
			ListOptions: gl.ListOptions{
				Page:    pageCounter,
				PerPage: perPage,
			},
			OrderBy: &orderBy,
			Sort:    &sortDir,
		})
		if err != nil {
			return nil, err
		}
		if len(tmpJobs) == 0 {
			break
		}
		foundOlder := false
		for _, job := range tmpJobs {
			/* if job.CreatedAt.Truncate(time.Hour * 24).Before(*startDate) {
				foundOlder = true
			} */
			createdAt := job.CreatedAt.Truncate(time.Hour * 24)
			if (createdAt.After(*startDate) || createdAt.Equal(*startDate)) && (createdAt.Before(*endDate) || createdAt.Equal(*endDate)) {
				//jobs = append(jobs, job)
				if _, ok := jobsMap[job.ID]; !ok {
					jobsMap[job.ID] = job
					jobIds = append(jobIds, job.ID)
				}
			} else {
				foundOlder = true
			}
		}
		if foundOlder {
			break nextPages
		}
		pageCounter++
	}
	fmt.Printf("Finished Getting remaining jobs")
	fmt.Printf("Starting sort of jobs")
	if len(jobsMap) > 0 {
		sort.Ints(jobIds)
		sort.Sort(sort.Reverse(sort.IntSlice(jobIds)))
		for _, id := range jobIds {
			jobs = append(jobs, jobsMap[id])
		}
	}
	fmt.Printf("Finished sorting of jobs")
	return jobs, nil
}

// Keep in mind, jobs are ordered descending by date, that means jobs[0].CreatedAt is newer
// than jobs[len(jobs) - 1].CreatedAt
// Todo: Improve by using a binary search if days > 1
// Keep in mind, jobs are ordered descending by date, that means jobs[0].CreatedAt is newer
// than jobs[len(jobs) - 1].CreatedAt
func HaveJobBetweenDates(jobs []*gl.Job, startDate time.Time, endDate time.Time) bool {
	if len(jobs) == 0 {
		return false
	}
	// making sure the dates are truncated to 24 hours so we only have the date to compare
	startDate = startDate.Truncate(time.Hour * 24)
	endDate = endDate.Truncate(time.Hour * 24)
	first := jobs[0].CreatedAt.Truncate(time.Hour * 24)
	last := jobs[len(jobs)-1].CreatedAt.Truncate(time.Hour * 24)
	days := int(first.Sub(last).Hours() / 24)
	if days == 0 {
		return (first.Equal(startDate) || first.After(startDate)) && (first.Equal(endDate) || first.Before(endDate))
	}
	if days == 1 {
		return ((first.Equal(startDate) || first.After(startDate)) &&
			(first.Equal(endDate) || first.Before(endDate))) || ((last.Equal(startDate) || last.After(startDate)) &&
			(last.Equal(endDate) || last.Before(endDate)))
	}
	if days > 1 {
		for _, job := range jobs {
			jobCreatedAt := job.CreatedAt.Truncate(time.Hour * 24)
			if (jobCreatedAt.Equal(startDate) || jobCreatedAt.After(startDate)) && (jobCreatedAt.Equal(endDate) || jobCreatedAt.Before(endDate)) {
				return true
			}
		}
	}
	return false
}

func GetClient(baseUrl string, token string) (*gl.Client, error) {
	client, err := gl.NewClient(token, gl.WithBaseURL(baseUrl))
	if err != nil {
		return nil, err
	}
	return client, nil
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
