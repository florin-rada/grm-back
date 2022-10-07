package login

import (
	"back/login"

	"github.com/gin-gonic/gin"
)

func CheckLoginMidleware() gin.HandlerFunc {
	return func(next *gin.Context) {
		tokenStruct := &login.PrivateTokenStruct{}
		err := next.ShouldBindHeader(&tokenStruct)
		if err != nil {
			next.JSON(500, gin.H{
				"error": "Error parsing header",
			})
			next.Abort()
			return
		}
		if isValid, err := login.ValidateToken(tokenStruct.PrivateToken); err == nil && isValid {
			next.Next()
			return
		} else {
			next.Abort()
			next.JSON(401, gin.H{
				"error": "Unauthorized",
			})
			return
		}
	}
}
