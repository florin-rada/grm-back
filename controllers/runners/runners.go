package runners

import (
	"net/http"
	"strconv"
	"time"

	consterrors "github.com/florin-rada/grm-back/const_errors"
	jobsModel "github.com/florin-rada/grm-back/models/jobs"
	"github.com/florin-rada/grm-back/models/synchronized"
	jobsservice "github.com/florin-rada/grm-back/services/jobs_service"
	runnersservice "github.com/florin-rada/grm-back/services/runners_service"
	"github.com/florin-rada/grm-back/utils"

	gl "gitlab.com/gitlab-org/api/client-go/v2"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type RunnerController struct {
	db *gorm.DB
	rs *runnersservice.RunnersService
	js *jobsservice.JobsService
}

func NewRunnerController(db *gorm.DB) *RunnerController {
	return &RunnerController{
		db: db,
		rs: runnersservice.NewRunnersService(runnersservice.NewRunnerRepository(db)),
		js: jobsservice.NewJobsService(jobsservice.NewJobsRepository(db), synchronized.NewSynchronizedRepository(db)),
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
	var page int64 = 1
	runners := []*gl.Runner{}

	for !foundAll {
		receivedRunners, _, err := rc.rs.GetAllRunners(gitClient, page, 100)
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
	runners, err := rc.rs.GetUserRunners(userID)
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
	err = rc.rs.DeleteRunner(runnerID, userID)
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
	err = rc.rs.AddRunnerForUser(gitClient, userID, runnerID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error":    err.Error(),
			"response": "",
		})
		return
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

	r, err := rc.rs.GetRunner(runnerID)
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
		MaximumTimeout *int64  `url:"maximum_timeout,omitempty" json:"maximum_timeout,omitempty"`
		MaxConcurrent  *int64  `url:"max_concurrent,omitempty" json:"max_concurrent,omitempty"`
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
	err = rc.rs.UpdateRunnerOnGit(client, r)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error":    err.Error(),
			"response": "",
		})
		return
	}
	err = rc.rs.UpdateRunner(r)
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
	r, err := rc.rs.GetRunner(runnerID)
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

	jobs, err := rc.js.SearchJobsForStatistics(userID, args)
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
