package runners

import (
	consterrors "back/const_errors"
	glModel "back/models/gitlab"
	jobsModel "back/models/jobs"
	"back/models/runners"
	"back/utils"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/xanzy/go-gitlab"
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
	runners := []*gitlab.Runner{}

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
	var numJobs int64
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
		tmpJobs, err := jm.GetRunnerJobs(uint(runnerID), userID, page, int(numJobs))
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
	args := struct {
		IdRunner int
	}{}
	err := ctx.BindJSON(&args)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error":    err.Error(),
			"response": "",
		})
		return
	}
	gitClient := utils.GetGitClientFromContext(ctx)
	if gitClient == nil {
		ctx.JSON(http.StatusOK, gin.H{
			"error":    consterrors.ErrNoGitToken.Error(),
			"response": "",
		})
		return
	}
	err = rc.rr.AddRunnerForUser(gitClient, userID, args.IdRunner)
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
