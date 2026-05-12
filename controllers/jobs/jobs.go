package jobs

import (
	"net/http"
	"strconv"
	"time"

	consterrors "github.com/florin-rada/grm-back/const_errors"
	jobsModel "github.com/florin-rada/grm-back/models/jobs"
	"github.com/florin-rada/grm-back/models/synchronized"
	jobsservice "github.com/florin-rada/grm-back/services/jobs_service"
	"github.com/florin-rada/grm-back/utils"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type JobsController struct {
	db *gorm.DB
	js *jobsservice.JobsService
}

func NewJobsController(db *gorm.DB) *JobsController {
	return &JobsController{db: db, js: jobsservice.NewJobsService(jobsservice.NewJobsRepository(db), synchronized.NewSynchronizedRepository(db))}
}

func (jc JobsController) ListRunnerJobs(ctx *gin.Context) {
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
	err = ctx.BindJSON(&params)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error":    consterrors.ErrInvalidSearchCriteria.Error(),
			"response": "",
		})
		return
	}

	jobs, err := jc.js.SearchRunnerJobs(runnerID, userID, params)
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

func (jc JobsController) TestGetJobsBetween(ctx *gin.Context) {
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
	jobs, err := jc.js.GetJobsBetween(gitClient, runnerID, &startDate, &endDate)
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

func (jc JobsController) GetLatestJobsForRunner(ctx *gin.Context) {
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
	for len(jobs) < int(numJobs) {
		tmpJobs, err := jc.js.GetRunnerJobs(uint(runnerID), userID, page, int(numJobs), "desc")
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
