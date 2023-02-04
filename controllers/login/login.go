package login

import (
	"back/models/keycloak"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

type AuthData struct {
	Username string `header:"username" json:"username"`
	Password string `header:"password" json:"password"`
}

type AccessToken struct {
	AccessToken string `header:"Private-Token"`
}

// This checks the token received from the app
func CheckLoginMidleware() gin.HandlerFunc {
	return func(next *gin.Context) {
		tokenStruct := &AccessToken{}
		err := next.ShouldBindHeader(&tokenStruct)
		if err != nil {
			next.JSON(500, gin.H{
				"error": "Error parsing header",
			})
			next.Abort()
			return
		}
		err = keycloak.ValidateToken(tokenStruct.AccessToken)
		if err != nil {
			next.JSON(http.StatusUnauthorized, gin.H{
				"error": err.Error(),
			})
			return
		}
		ui, err := keycloak.GetUserInfo(tokenStruct.AccessToken)
		if err != nil {
			next.JSON(http.StatusUnauthorized, gin.H{
				"error": err.Error(),
			})
			return
		}
		//next.Set("user_git_token", "asdg12398as971982jckalsuyu182")
		//next.Set("user_info", ui)
		next.Set("user_id", *ui.Sub)
		next.Set("username", *ui.PreferredUsername)
		next.Set("email", *ui.Email)
		next.Next()
	}
}

func ValidateLogin(ctx *gin.Context) {
	fmt.Printf("Trying to login")
	authArgs := AuthData{}
	err := ctx.BindJSON(&authArgs)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}
	fmt.Printf("We have arguments: %+v", authArgs)
	accessToken, err := keycloak.LoginUser(authArgs.Username, authArgs.Password)
	if err != nil {
		fmt.Printf("We have unauthorized status, error: %s", err.Error())
		ctx.JSON(http.StatusUnauthorized, gin.H{
			"error": err.Error(),
		})
		return
	}
	fmt.Printf("We have valid credentials and received access token: %+v", &accessToken)
	ctx.JSON(http.StatusOK, gin.H{
		"error":         "",
		"access_token":  accessToken.AccessToken,
		"refresh_token": accessToken.RefreshToken,
		"expires_in":    accessToken.ExpiresIn,
	})
}

/* func CheckLoginMidleware() gin.HandlerFunc {
	return func(next *gin.Context) {

		next.Next()
		return
	}
} */
