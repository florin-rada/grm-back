package auth

import (
	"net/http"

	"github.com/florin-rada/grm-back/config"
	auth_service "github.com/florin-rada/grm-back/services/auth_service"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type AuthController struct {
	db          *gorm.DB
	Config      *config.AuthConfig
	authService *auth_service.AuthService
}

func NewAuthController(db *gorm.DB, config *config.AuthConfig) *AuthController {
	return &AuthController{db: db, Config: config, authService: auth_service.NewAuthService(db, config)}
}

func (ac *AuthController) RedirectToGitlabAuth(ctx *gin.Context) {
	sess := sessions.Default(ctx)
	state := ac.authService.GenerateOauth2State(16) // Generate a random state string for CSRF protection
	verifier := ac.authService.GetOauth2Verifier()
	//state := randomString(16)   // Generate a random state string for CSRF protection
	//ac.authService.GetGitlabAuthURL(state, verifier)
	authURL, err := ac.authService.GetGitlabAuthURL(state, verifier)
	if err != nil {
		ctx.JSON(500, gin.H{
			"error": err.Error(),
		})
		return
	}
	sess.Set("oauth_state", state)
	sess.Set("oauth_verifier", verifier)
	sess.Save()
	ctx.Redirect(http.StatusTemporaryRedirect, authURL)
}

func (ac *AuthController) GitlabAuthCallback(ctx *gin.Context) {
	sess := sessions.Default(ctx)
	storedState := sess.Get("oauth_state")
	storedVerifier := sess.Get("oauth_verifier")
	if storedState == nil || storedVerifier == nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "State or PKCE verifier not found in session",
		})
		return
	}
	state := ctx.Query("state")
	code := ctx.Query("code")
	if state != storedState.(string) {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid state parameter",
		})
		return
	}
	token, err := ac.authService.HandleGitlabAuthCallback(ctx, code, storedVerifier.(string))
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}
	sess.Delete("oauth_state")
	sess.Delete("oauth_verifier")
	sess.Set("access_token", token.AccessToken)
	sess.Set("refresh_token", token.RefreshToken)
	sess.Save()
	ctx.JSON(http.StatusOK, gin.H{
		"error":         "",
		"access_token":  token.AccessToken,
		"refresh_token": token.RefreshToken,
	})

}
