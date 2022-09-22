package login

import (
	"github.com/gin-gonic/gin"
)

/* type AuthData struct {
	Username string `header:"username"`
	Password string `header:"password"`
} */

type PrivateTokenStruct struct {
	PrivateToken string `header:"private_token"`
}

func CheckLoginMidleware() gin.HandlerFunc {
	return func(next *gin.Context) {
		tokenStruct := &PrivateTokenStruct{}
		err := next.ShouldBindHeader(&tokenStruct)
		if err != nil {
			next.JSON(500, gin.H{
				"error": "Error parsing header",
			})
			next.Abort()
			return
		}
		if isValid, err := validateToken(tokenStruct.PrivateToken); err == nil && isValid {
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
	/* return func(next *gin.Context) {
		authData := AuthData{}
		err := next.ShouldBindHeader(&authData)
		if err != nil {
			next.JSON(500, gin.H{
				"error": "Error parsing header",
			})
			return
		}

		if isValid, err := validateLogin(authData.Username, authData.Password); err == nil && isValid {
			next.Next()
			return
		} else {
			next.Abort()
			next.JSON(401, gin.H{
				"error": "Unauthorized",
			})
		}
	} */
}

func validateToken(token string) (bool, error) {
	return true, nil
}

func validateLogin(user string, pass string) (bool, error) {
	if user == "user" && pass == "pass" {
		return true, nil
	}
	return false, nil
}

func init() {

}
