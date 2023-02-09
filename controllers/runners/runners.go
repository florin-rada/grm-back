package runners

import (
	glModel "back/models/gitlab"
	"back/models/runners"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/xanzy/go-gitlab"
)

func ListUserRunnersFromGit(ctx *gin.Context) {
	gitClientI, exists := ctx.Get("git_client")
	if !exists {
		ctx.JSON(http.StatusOK, gin.H{
			"error":    "No git token",
			"response": "",
		})
		return
	}
	gitClient := gitClientI.(*gitlab.Client)
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
	userIDI, exists := ctx.Get("user_id")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{
			"error":    "Not logged in",
			"response": "",
		})
		return
	}
	userID := userIDI.(string)
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
	/* userIDI, exists := ctx.Get("user_id")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{
			"error":    "Not logged in",
			"response": "",
		})
		return
	}
	userID := userIDI.(string) */
}
