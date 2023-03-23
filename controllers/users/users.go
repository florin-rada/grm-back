package users

import (
	consterrors "back/const_errors"
	"back/models/gitlab"
	"back/models/keycloak"
	"back/models/users"
	"back/utils"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
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
			"error": consterrors.ErrNoToken.Error(),
		})
		return
	}
	err = keycloak.ValidateToken(ts.AccessToken)
	if err != nil {
		ctx.JSON(http.StatusForbidden, gin.H{
			"error": consterrors.ErrInvalidToken.Error(),
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
			"error": consterrors.ErrNoToken.Error(),
		})
		return
	}
	jwt, err := keycloak.RefreshToken(ts.AccessToken)
	if err != nil {
		ctx.JSON(http.StatusForbidden, gin.H{
			"error": consterrors.ErrInvalidToken.Error(),
		})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{
		"error":        "",
		"access_token": jwt.AccessToken,
	})
}

// we make sure to hide the git token. The git token must never be exposed
func GetUserGitDetails(ctx *gin.Context) {
	userID := utils.GetUserIdFromContext(ctx)
	if userID == "" {
		ctx.JSON(http.StatusUnauthorized, gin.H{
			"error": consterrors.ErrNotLoggedIn.Error(),
		})
		return
	}

	gld, err := users.GetUserGitDetails(userID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}
	// we hide the token
	// the git token must never be returned in the application as it is not used in the frontend
	gld.Token = ""
	ctx.JSON(http.StatusOK, gin.H{
		"error":    "",
		"response": gld,
	})
}

func UpdateUserGitDetails(ctx *gin.Context) {
	userID := utils.GetUserIdFromContext(ctx)
	if userID == "" {
		ctx.JSON(http.StatusUnauthorized, gin.H{
			"error": consterrors.ErrNotLoggedIn.Error(),
		})
		return
	}
	gld, err := users.GetUserGitDetails(userID)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}
	if errors.Is(err, gorm.ErrRecordNotFound) {
		gld = &users.GitlabDetails{
			IDUser: userID,
		}
	}
	args := struct {
		InstanceURL string `json:"instance_url"`
		Token       string `json:"token"`
		SyncRate    int    `json:"sync_rage"`
	}{}

	err = ctx.BindJSON(&args)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}
	gld.InstanceURL = args.InstanceURL
	gld.Token = args.Token
	gld.SyncRate = args.SyncRate
	err = users.SetUserGitDetails(gld)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{
		"error": "",
	})
}

func TestGitConnection(ctx *gin.Context) {
	args := struct {
		InstanceURL string `json:"instance_url"`
		Token       string `json:"token"`
	}{}
	err := ctx.BindJSON(&args)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}
	gitClient, err := gitlab.GetClient(args.InstanceURL, args.Token)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "error connecting to gitlab",
		})
		return
	}
	_, _, err = gitlab.GetAllRunners(gitClient, 1, 100)
	if err != nil {
		ctx.JSON(http.StatusForbidden, gin.H{
			"error": "token is not authorized",
		})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{
		"error":    "",
		"response": "OK",
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
