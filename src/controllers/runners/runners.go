package runners

import (
	consterrors "back/pkg/const_errors"
	glModel "back/pkg/models/gitlab"
	jobsModel "back/pkg/models/jobs"
	"back/pkg/models/runners"
	"back/pkg/utils"
	"net/http"
	"strconv"
	"time"

	gl "gitlab.com/gitlab-org/api/client-go/v2"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type RunnerController struct {
	db *gorm.DB
	rr *runners.RunnerRepository
}

func NewRunnerController(db *gorm.DB) *RunnerController {
	return &RunnerController{
		db: db,
		rr: runners.NewRunnerRepository(db),
	}
}

func (rc RunnerController) ListUserRunnersFromGit(ctx *gin.Context) {
	gitClient := utils.GetGitClientFromContext(ctx)
	if gitClient == nil {
		ctx.JSON(http.StatusOK, gin.H{
			"error":    consterrors.ErrNoGitToken.Error(),
			"response": "",
		})
		return
	}
	var foundAll bool
	var page int = 1
	runners := []*gl.Runner{}

	for !foundAll {
		receivedRunners, _, err := glModel.GetAllRunners(gitClient, page, 100)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{
				"error":    err.Error(),
				"response": "",
			})
		}
		runners = append(runners, receivedRunners...)
		if len(receivedRunners) == 0 || len(receivedRunners) < 100 {
			foundAll = true
			break
		}
	}
	ctx.JSON(http.StatusOK, gin.H{
		"error":    "",
		"response": runners,
	})
}

func (rc RunnerController) ListUserRunners(ctx *gin.Context) {
	userID := utils.GetUserIdFromContext(ctx)
	if userID == "" {
		ctx.JSON(http.StatusUnauthorized, gin.H{
			"error":    consterrors.ErrNotLoggedIn.Error(),
			"response": "",
		})
		return
	}
	runners, err := rc.rr.GetUserRunners(userID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error":    err.Error(),
			"response": "",
		})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{
		"error":    "",
		"response": runners,
	})
}

func (rc RunnerController) DeleteRunner(ctx *gin.Context) {
	userID := utils.GetUserIdFromContext(ctx)
	if userID == "" {
		ctx.JSON(http.StatusUnauthorized, gin.H{
			"error":    consterrors.ErrNotLoggedIn.Error(),
			"response": "",
		})
		return
	}
	client := utils.GetGitClientFromContext(ctx)
	if client == nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error":    consterrors.ErrNoGitClient,
			"response": "",
		})
		return
	}
	runnerIDStr := ctx.Param("id")
	if runnerIDStr == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error":    consterrors.ErrNoID.Error(),
			"response": "",
		})
		return
	}
	runnerID, err := strconv.ParseInt(runnerIDStr, 10, 64)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error":    consterrors.ErrNoID,
			"response": "",
		})
		return
	}
	err = rc.rr.DeleteRunner(uint(runnerID), userID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error":    err.Error(),
			"response": "",
		})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{
		"error":    "",
		"response": "OK",
	})

}

func (rc RunnerController) ListRunnerJobs(ctx *gin.Context) {
	userID := utils.GetUserIdFromContext(ctx)
	if userID == "" {
		ctx.JSON(http.StatusUnauthorized, gin.H{
			"error":    consterrors.ErrNotLoggedIn.Error(),
			"response": "",
		})
		return
	}
	client := utils.GetGitClientFromContext(ctx)
	if client == nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error":    consterrors.ErrNoGitClient,
			"response": "",
		})
		return
	}
	runnerIDStr := ctx.Param("id")
	if runnerIDStr == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error":    consterrors.ErrNoID.Error(),
			"response": "",
		})
		return
	}
	runnerID, err := strconv.ParseInt(runnerIDStr, 10, 64)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error":    consterrors.ErrNoID,
			"response": "",
		})
		return
	}
	params := jobsModel.JobSearchArgs{}
	err = ctx.BindJSON(params)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error":    consterrors.ErrInvalidSearchCriteria.Error(),
			"response": "",
		})
		return
	}
	jm := jobsModel.NewJobsModel(rc.db, client)
	jobs, err := jm.SearchRunnerJobs(uint(runnerID), userID, params)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error":    err.Error(),
			"response": "",
		})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{
		"error":    "",
		"response": jobs,
	})
}

func (rc RunnerController) GetLatestJobsForRunner(ctx *gin.Context) {
	userID := utils.GetUserIdFromContext(ctx)
	if userID == "" {
		ctx.JSON(http.StatusUnauthorized, gin.H{
			"error":    consterrors.ErrNotLoggedIn.Error(),
			"response": "",
		})
		return
	}
	client := utils.GetGitClientFromContext(ctx)
	if client == nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error":    consterrors.ErrNoGitClient,
			"response": "",
		})
		return
	}
	runnerIDStr := ctx.Param("id")
	if runnerIDStr == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error":    consterrors.ErrNoID.Error(),
			"response": "",
		})
		return
	}
	runnerID, err := strconv.ParseInt(runnerIDStr, 10, 64)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error":    consterrors.ErrNoID,
			"response": "",
		})
		return
	}
	var numJobs int64 = 25
	numJobsStr := ctx.Query("num_jobs")
	if numJobsStr != "" {
		numJobs, err = strconv.ParseInt(numJobsStr, 10, 64)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{
				"error":    consterrors.ErrInvalidNumJobs,
				"response": "",
			})
			return
		}
	}
	jobs := []jobsModel.Job{}
	page := 0
	jm := jobsModel.NewJobsModel(rc.db, client)
	for len(jobs) < int(numJobs) {
		tmpJobs, err := jm.GetRunnerJobs(uint(runnerID), userID, page, int(numJobs), "desc")
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{
				"error":    err.Error(),
				"response": "",
			})
			return
		}
		if len(tmpJobs) == 0 {
			break
		}
		jobs = append(jobs, tmpJobs...)
		page++
	}

	ctx.JSON(http.StatusOK, gin.H{
		"error":    "",
		"response": jobs,
	})
}

func (rc RunnerController) AddRunnerForUser(ctx *gin.Context) {
	userID := utils.GetUserIdFromContext(ctx)
	if userID == "" {
		ctx.JSON(http.StatusUnauthorized, gin.H{
			"error":    consterrors.ErrNotLoggedIn.Error(),
			"response": "",
		})
		return
	}
	runnerIDstr := ctx.Param("id")
	if runnerIDstr == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error":    consterrors.ErrNoID.Error(),
			"response": "",
		})
		return
	}
	runnerID, err := strconv.ParseInt(runnerIDstr, 10, 64)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error":    consterrors.ErrInvalidID,
			"response": "",
		})
		return
	}
	/* args := struct {
		IdRunner int `json:"git_id_runner"`
	}{}
	err := ctx.BindJSON(&args)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error":    err.Error(),
			"response": "",
		})
		return
	} */
	gitClient := utils.GetGitClientFromContext(ctx)
	if gitClient == nil {
		ctx.JSON(http.StatusOK, gin.H{
			"error":    consterrors.ErrNoGitToken.Error(),
			"response": "",
		})
		return
	}
	err = rc.rr.AddRunnerForUser(gitClient, userID, int(runnerID))
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error":    err.Error(),
			"response": "",
		})
	}
	ctx.JSON(http.StatusOK, gin.H{
		"error":    "",
		"response": "",
	})
}

func (rc RunnerController) UpdateRunner(ctx *gin.Context) {
	client := utils.GetGitClientFromContext(ctx)
	if client == nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error":    consterrors.ErrNoGitClient,
			"response": "",
		})
		return
	}
	userID := utils.GetUserIdFromContext(ctx)
	if userID == "" {
		ctx.JSON(http.StatusUnauthorized, gin.H{
			"error":    consterrors.ErrNotLoggedIn.Error(),
			"response": "",
		})
		return
	}
	runnerIDStr := ctx.Param("id")
	if runnerIDStr == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error":    consterrors.ErrNoID.Error(),
			"response": "",
		})
		return
	}
	runnerID, err := strconv.ParseInt(runnerIDStr, 10, 64)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error":    consterrors.ErrInvalidID.Error(),
			"response": "",
		})
		return
	}

	r, err := rc.rr.GetRunner(uint(runnerID))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error":    err.Error(),
			"response": "",
		})
		return
	}

	args := struct {
		Description    *string `url:"description,omitempty" json:"description,omitempty"`
		Paused         *bool   `url:"paused,omitempty" json:"paused,omitempty"`
		TagList        *string `url:"tag_list,omitempty" json:"tag_list,omitempty"`
		RunUntagged    *bool   `url:"run_untagged,omitempty" json:"run_untagged,omitempty"`
		MaximumTimeout *int    `url:"maximum_timeout,omitempty" json:"maximum_timeout,omitempty"`
		MaxConcurrent  *int    `url:"max_concurrent,omitempty" json:"max_concurrent,omitempty"`
		Locked         *bool   `url:"locked,omitempty" json:"locked,omitempty"`
	}{}
	err = ctx.BindJSON(&args)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error":    err.Error(),
			"response": "",
		})
		return
	}
	if args.Description != nil {
		r.Description = *args.Description
	}
	if args.Paused != nil {
		r.Paused = *args.Paused
	}
	if args.TagList != nil {
		r.TagList = *args.TagList
	}
	if args.RunUntagged != nil {
		r.RunUntagged = *args.RunUntagged
	}
	if args.MaximumTimeout != nil {
		r.MaximumTimeout = *args.MaximumTimeout
	}
	if args.MaxConcurrent != nil {
		r.MaxConcurrent = *args.MaxConcurrent
	}

	if args.Locked != nil {
		r.Locked = *args.Locked
	}
	err = rc.rr.UpdateRunnerOnGit(client, r)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error":    err.Error(),
			"response": "",
		})
		return
	}
	err = rc.rr.UpdateRunner(r)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error":    err.Error(),
			"response": "",
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"error":    "",
		"response": "OK",
	})

}

func (rc RunnerController) GetRunnerDetails(ctx *gin.Context) {
	userID := utils.GetUserIdFromContext(ctx)
	if userID == "" {
		ctx.JSON(http.StatusUnauthorized, gin.H{
			"error":    consterrors.ErrNotLoggedIn.Error(),
			"response": "",
		})
		return
	}
	runnerIDStr := ctx.Param("id")
	if runnerIDStr == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error":    consterrors.ErrNoID.Error(),
			"response": "",
		})
		return
	}
	runnerID, err := strconv.ParseInt(runnerIDStr, 10, 64)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error":    consterrors.ErrInvalidID.Error(),
			"response": "",
		})
		return
	}
	r, err := rc.rr.GetRunner(uint(runnerID))
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error":    err.Error(),
			"response": "",
		})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{
		"error":    "",
		"response": r,
	})
}

func (rc RunnerController) GetStatisticsData(ctx *gin.Context) {
	gitClient := utils.GetGitClientFromContext(ctx)
	if gitClient == nil {
		ctx.JSON(http.StatusOK, gin.H{
			"error":    consterrors.ErrNoGitToken.Error(),
			"response": "",
		})
		return
	}

	userID := utils.GetUserIdFromContext(ctx)
	if userID == "" {
		ctx.JSON(http.StatusUnauthorized, gin.H{
			"error":    consterrors.ErrNotLoggedIn.Error(),
			"response": "",
		})
		return
	}
	args := jobsModel.StatisticsSearchArgs{}
	err := ctx.BindJSON(&args)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error":    err.Error(),
			"response": "",
		})
		return
	}
	truncated := args.CreatedAtStart.In(time.UTC).Truncate(time.Hour * 24)
	args.CreatedAtStart = &truncated
	truncated = args.CreatedAtEnd.In(time.UTC).Truncate(time.Hour * 24)
	args.CreatedAtEnd = &truncated

	if args.CreatedAtStart.After(time.Now()) {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error":    "date after today",
			"response": "",
		})
		return
	}
	if args.CreatedAtEnd.After(time.Now()) {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error":    "date after today",
			"response": "",
		})
		return
	}
	jm := jobsModel.NewJobsModel(rc.db, gitClient)

	jobs, err := jm.SearchJobsForStatistics(userID, args)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error":    err.Error(),
			"response": "",
		})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{
		"error":    "",
		"response": jobs,
	})
}

func (rc RunnerController) TestGetJobsBetween(ctx *gin.Context) {
	gitClient := utils.GetGitClientFromContext(ctx)
	if gitClient == nil {
		ctx.JSON(http.StatusOK, gin.H{
			"error":    consterrors.ErrNoGitToken.Error(),
			"response": "",
		})
		return
	}

	userID := utils.GetUserIdFromContext(ctx)
	if userID == "" {
		ctx.JSON(http.StatusUnauthorized, gin.H{
			"error":    consterrors.ErrNotLoggedIn.Error(),
			"response": "",
		})
		return
	}

	startDate, err := time.Parse("2006-01-02", ctx.Query("start_date"))
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error":    err.Error(),
			"response": "",
		})
		return
	}
	endDate, err := time.Parse("2006-01-02", ctx.Query("end_date"))
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error":    err.Error(),
			"response": "",
		})
		return
	}
	runnerID, err := strconv.ParseInt(ctx.Query("runner_id"), 10, 64)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error":    err.Error(),
			"response": "",
		})
		return
	}
	jobs, err := glModel.GetJobsBetween(gitClient, int(runnerID), &startDate, &endDate)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error":    err.Error(),
			"response": "",
		})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{
		"error":    "",
		"response": jobs,
	})
}
