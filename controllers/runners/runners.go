package runners

import (
	consterrors "back/const_errors"
	glModel "back/models/gitlab"
	"back/models/runners"
	"back/utils"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/xanzy/go-gitlab"
)

func ListUserRunnersFromGit(ctx *gin.Context) {
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

func ListUserRunners(ctx *gin.Context) {
	userID := utils.GetUserIdFromContext(ctx)
	if userID == "" {
		ctx.JSON(http.StatusUnauthorized, gin.H{
			"error":    consterrors.ErrNotLoggedIn.Error(),
			"response": "",
		})
		return
	}
	runners, err := runners.GetUserRunners(userID)
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

func AddRunnerForUser(ctx *gin.Context) {
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
	err = runners.AddRunnerForUser(gitClient, userID, args.IdRunner)
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
