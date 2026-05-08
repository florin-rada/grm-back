package utils

import (
	gl "gitlab.com/gitlab-org/api/client-go/v2"

	"github.com/gin-gonic/gin"
)

func GetUserIdFromContext(ctx *gin.Context) string {
	userIDI, exists := ctx.Get("user_id")
	if !exists {
		return ""
	}
	userID := userIDI.(string)
	return userID
}

func GetGitClientFromContext(ctx *gin.Context) *gl.Client {
	gitClientI, exists := ctx.Get("git_client")
	if !exists {
		return nil
	}
	gitClient := gitClientI.(*gl.Client)
	return gitClient
}
