package jobsservice

import (
	"errors"
	"fmt"
	"sort"
	"strconv"
	"time"

	consterrors "github.com/florin-rada/grm-back/const_errors"
	"github.com/florin-rada/grm-back/mapper"
	jobs_model "github.com/florin-rada/grm-back/models/jobs"
	"github.com/florin-rada/grm-back/models/projects"
	"github.com/florin-rada/grm-back/models/synchronized"
	gl "gitlab.com/gitlab-org/api/client-go/v2"
	"gorm.io/gorm/clause"
)

type JobsService struct {
	repo     *JobsRepository
	syncRepo *synchronized.SynchronizedRepository
}

func NewJobsService(jr *JobsRepository, syncRepo *synchronized.SynchronizedRepository) *JobsService {
	return &JobsService{repo: jr, syncRepo: syncRepo}
}

func (js JobsService) SyncRunnerJobs(client *gl.Client, userID string, runnerID int64) error {
	fmt.Printf("Starting updating runner jobs for runner %d\n", runnerID)
	glJobs, _, err := js.GetRunnerJobsFromGit(client, int(runnerID), "", 0, 20)
	if err != nil {
		return err
	}
	prjs := []projects.Project{}
	jobs := make([]jobs_model.Job, 0, len(glJobs))
	for _, job := range glJobs {
		translated, err := mapper.TranslateGLJobToJob(userID, job)
		if err != nil {
			return err
		}
		translated.RunnerID = int64(runnerID)
		jobs = append(jobs, *translated)
		prj := projects.Project{
			ID:             job.Project.ID,
			Name:           job.Project.Name,
			InternalUserID: userID,
		}
		prjs = append(prjs, prj)
	}
	if err := js.repo.UpsertJobs(jobs); err != nil {
		return err
	}
	// this is bad, either need to create a system to notify projectsModel or move the sync section to a stand alone package
	resp := js.repo.db.Clauses(clause.OnConflict{
		UpdateAll: true,
	}).Create(prjs)
	if resp.Error != nil {
		return resp.Error
	}
	fmt.Printf("Ending updating runner jobs for runner %d\n", runnerID)
	return nil
}

func (js JobsService) SyncJobsBetweenDates(client *gl.Client, userID string, runnerID int64, startDate *time.Time, endDate *time.Time) error {
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
	notSyncedStartDate, notSyncedEndDate, err := js.syncRepo.GetMinMaxUnsyncedDates(userID, runnerID, &truncatedStartDate, &truncatedEndDate)
	if err != nil {
		return err
	}
	if notSyncedStartDate == nil && notSyncedEndDate == nil {
		return nil
	}

	gljobs, err := js.GetJobsBetween(client, runnerID, notSyncedStartDate, notSyncedEndDate)
	if err != nil {
		return err
	}
	if len(gljobs) == 0 {
		return nil
	}
	jobs := make([]jobs_model.Job, 0, len(gljobs))
	for _, job := range gljobs {
		translated, err := mapper.TranslateGLJobToJob(userID, job)
		if err != nil {
			return err
		}
		jobs = append(jobs, *translated)
	}
	if err := js.repo.UpsertJobs(jobs); err != nil {
		return err
	}
	err = js.syncRepo.SaveSyncedDates(userID, runnerID, *notSyncedStartDate, *notSyncedEndDate)
	if err != nil {
		return err
	}
	return nil
}

func (js JobsService) GetJobsBetween(client *gl.Client, runnerID int64, startDate *time.Time, endDate *time.Time) ([]*gl.Job, error) {
	if client == nil {
		return nil, consterrors.ErrNoGitClient
	}
	jobs := []*gl.Job{}
	// due to moving nature of the results returned from the api
	// sometimes the same job can apear in two results
	// to eliminate this we store the jobs in a map with the ID as the key
	// and the ids in a slice
	// this way at the end we can sort the ID slice and than add them to our slice result
	jobsMap := make(map[int64]*gl.Job)
	jobIds := []int64{}
	if startDate == nil || endDate == nil {
		return []*gl.Job{}, errors.New("invalid start date or end date")
	}

	/* if endDate.Before(*startDate) {
		return []*gl.Job{}, errors.New("start date is before end date")
	} */
	perPage := int64(100)
	// we make sure our start and end date are truncated to only 24 hours
	tmpStartDate := startDate.Truncate(24 * time.Hour)
	startDate = &tmpStartDate
	tmpEndDate := endDate.Truncate(24 * time.Hour)
	endDate = &tmpEndDate
	fmt.Printf("Start Date: %v\nEnd Date: %v", *startDate, *endDate)
	//foundEveryting := false
	// we calculate how many days are between now and our end date
	//numDaysSinceEndDate := int64(time.Since(*endDate) / (24 * time.Hour))
	//var pagesInFirstDay int
	var page int64 = 1
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
	if len(tmpJobs) < int(perPage) {
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
	var maxPage int64
	if maxPageStr != "" {
		tmp, err := strconv.ParseInt(maxPageStr, 10, 64)
		if err != nil {
			fmt.Printf("Warning, no x-total-pages header received, skipping\n")
		} else {
			maxPage = tmp
		}
	}
	fmt.Printf("Date of last job in tmpJobs: %v", tmpJobs[len(tmpJobs)-1].CreatedAt.Truncate(time.Hour*24))
	fmt.Printf("Num pages as received from x-total-page: %d", maxPage)
	// we know we don't have anything on the first page so we try to find our desired pages
	// try to find max limit by multiplying the number of days between today and end date
	// with a counter until we either receive a empty list or a date older than end date
	//minPage := 1
	if maxPage == 0 {
		maxPage, err = js.GetJobsMaxPageForDate(client, runnerID, *endDate)
		if err != nil {
			return nil, err
		}
	}
	fmt.Printf("Starting binary search for our targeted dates")

	maxPage, err = js.GetJobsTargetPage(client, runnerID, maxPage, *startDate, *endDate)
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
		sort.Slice(jobIds, func(i, j int) bool { return jobIds[i] > jobIds[j] })
		//sort.Sort(sort.Reverse(jobIds))
		for _, id := range jobIds {
			jobs = append(jobs, jobsMap[id])
		}
	}
	fmt.Printf("Finished sorting of jobs")
	return jobs, nil
}

// GetJobsTargetPage performs a binary search to find the page that has jobs
// with CreatedAt between our start and end dates
// It will return a page number that contains at least a job between
// our target start and end dates
// GetJobsTargetPage performs a binary search to find the page that has jobs
// with CreatedAt between our start and end dates
// It will return a page number that contains at least a job between
// our target start and end dates
func (js JobsService) GetJobsTargetPage(client *gl.Client, runnerID int64, maxPage int64, startDate time.Time, endDate time.Time) (int64, error) {
	fmt.Printf("GetJobsTargetPage started for RunnerID: %d, startDate: %v, endDate: %v", runnerID, startDate, endDate)
	defer fmt.Printf("GetJobsTargetPage finished for RunnerID: %d, startDate: %v, endDate: %v", runnerID, startDate, endDate)
	minPage := int64(1)
	orderBy := "id"
	sortDirection := "desc"
	prevMaxPage := minPage
	prevMinPage := maxPage
	for minPage <= maxPage {
		currentPage := (maxPage + minPage) / 2
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
		if js.HaveJobBetweenDates(tmpJobs, startDate, endDate) {
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

// Keep in mind, jobs are ordered descending by date, that means jobs[0].CreatedAt is newer
// than jobs[len(jobs) - 1].CreatedAt
// Todo: Improve by using a binary search if days > 1
// Keep in mind, jobs are ordered descending by date, that means jobs[0].CreatedAt is newer
// than jobs[len(jobs) - 1].CreatedAt
func (js JobsService) HaveJobBetweenDates(jobs []*gl.Job, startDate time.Time, endDate time.Time) bool {
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

// GetJobsMaxPageForDate will try and determine what is the max page for
// our binary search. It will iterate the number of days times a counter
// in it either finds a empty page or a page with jobs older than our
// target date
func (js JobsService) GetJobsMaxPageForDate(client *gl.Client, runnerID int64, endDate time.Time) (int64, error) {
	fmt.Printf("\nGetJobsMaxPageForDate: Starting for RunnerID: %d endDate: %v", runnerID, endDate)
	defer fmt.Printf("\nGetJobsMaxPageForDate: Ended for RunnerID: %d, endDate: %v", runnerID, endDate)
	pageCounter := int64(time.Since(endDate) / (24 * time.Hour))
	perPage := int64(100)
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

func (js JobsService) GetRunnerJobsFromGit(client *gl.Client, runnerID int, status string, page int64, perPage int64) ([]*gl.Job, *gl.Response, error) {
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
func GetJobsSince(client *gl.Client, runnerID int64, startDate *time.Time) ([]*gl.Job, error) {
	if client == nil {
		return nil, consterrors.ErrInvalidGitClient
	}
	jobs := []*gl.Job{}
	page := int64(0)
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

func (js JobsService) GetRunnerJobs(idRunner uint, idUser string, page int, perPage int, order string) ([]jobs_model.Job, error) {
	return js.repo.GetRunnerJobs(idRunner, idUser, page, perPage, order)
}

func (js JobsService) SearchRunnerJobs(idRunner int64, idUser string, params jobs_model.JobSearchArgs) ([]jobs_model.Job, error) {
	return js.repo.SearchRunnerJobs(idRunner, idUser, params)
}

func (js JobsService) SearchJobsForStatistics(idUser string, params jobs_model.StatisticsSearchArgs) ([]jobs_model.Job, error) {
	return js.repo.SearchJobsForStatistics(idUser, params)
}
