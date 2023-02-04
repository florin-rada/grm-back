package users

import (
	"back/models/keycloak"
	"back/models/users"
	"net/http"

	"github.com/gin-gonic/gin"
)

type tokenStruct struct {
	AccessToken string `json:"access_token"`
}

func Register(ctx *gin.Context) {
	args := struct {
		Email     string `json:"email" binding:"required"`
		Password  string `json:"password" binding:"required"`
		FirstName string `json:"first_name" binding:"required"`
		LastName  string `json:"last_name" binding:"required"`
		/*Country   string `json:"country" binding:"required"`
		City      string `json:"city" binding:"required"`
		Address   string `json:"address" binding:"required"` */
	}{}

	err := ctx.BindJSON(&args)
	if err != nil {
		ctx.JSON(500, gin.H{
			"error": err.Error(),
		})
	}

	u, err := users.CreateUser(args.Email, args.Password, args.FirstName, args.LastName)
	if err != nil {
		ctx.JSON(500, gin.H{
			"error": err.Error(),
		})
		return
	}

	ctx.JSON(200, gin.H{
		"error": "",
		"user":  u,
	})

}

func ValidateToken(ctx *gin.Context) {
	ts := tokenStruct{}
	err := ctx.ShouldBindHeader(&ts)
	if err != nil {
		ctx.JSON(http.StatusForbidden, gin.H{
			"error": "no token",
		})
		return
	}
	err = keycloak.ValidateToken(ts.AccessToken)
	if err != nil {
		ctx.JSON(http.StatusForbidden, gin.H{
			"error": "invalid token",
		})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{
		"error": "",
	})
}

func RefreshToken(ctx *gin.Context) {
	ts := tokenStruct{}
	err := ctx.ShouldBindHeader(&ts)
	if err != nil {
		ctx.JSON(http.StatusForbidden, gin.H{
			"error": "no token",
		})
		return
	}
	jwt, err := keycloak.RefreshToken(ts.AccessToken)
	if err != nil {
		ctx.JSON(http.StatusForbidden, gin.H{
			"error": "invalid token",
		})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{
		"error":        "",
		"access_token": jwt.AccessToken,
	})
}

// not needed anymore, email confirmation handled by keycloak
/* func ConfirmRegistration(ctx *gin.Context) {
	token := ctx.Query("token")
	if token == "" {
		ctx.JSON(400, gin.H{
			"error": errors.New("Missing token"),
		})
		return
	}
	err := users.ConfirmRegistration(db.PrivateDB, token)
	if err != nil {
		ctx.JSON(500, gin.H{
			"error": err.Error(),
		})
		return
	}
	ctx.JSON(200, gin.H{
		"error": "",
	})

} */
