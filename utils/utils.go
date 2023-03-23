package utils

import (
	"github.com/xanzy/go-gitlab"

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

func GetGitClientFromContext(ctx *gin.Context) *gitlab.Client {
	gitClientI, exists := ctx.Get("git_client")
	if !exists {
		return nil
	}
	gitClient := gitClientI.(*gitlab.Client)
	return gitClient
}
