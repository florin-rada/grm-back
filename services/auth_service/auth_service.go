package authservice

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"time"

	"github.com/florin-rada/grm-back/config"
	consterrors "github.com/florin-rada/grm-back/const_errors"
	gl "gitlab.com/gitlab-org/api/client-go/v2"
	"golang.org/x/oauth2"
	"gorm.io/gorm"
)

type AuthService struct {
	ar     *AuthRepository
	config *config.AuthConfig
	oauth  *oauth2.Config
}

func NewAuthService(db *gorm.DB, config *config.AuthConfig) *AuthService {
	return &AuthService{
		ar: NewAuthRepository(db, config), config: config,
		oauth: &oauth2.Config{
			ClientID:     config.GitlabAPIKey,
			ClientSecret: config.GitlabAPISecretKey,
			RedirectURL:  config.RedirectURL,
			Scopes:       []string{"read_user", "read_api", "read_repository"},
			Endpoint: oauth2.Endpoint{
				AuthURL:  config.GitlabAuthURL,
				TokenURL: config.GitlabTokenURL,
			},
		},
	}
}

// GetGitlabAuthURL generates the GitLab OAuth2 authorization URL with the provided state and code verifier for PKCE.
// Both the state and code verifier are used to enhance security during the OAuth2 flow, preventing CSRF attacks and ensuring the integrity of the authorization process.
// Both need to be present
func (as *AuthService) GetGitlabAuthURL(state string, verifier string) (string, error) {
	if state == "" || verifier == "" {
		return "", consterrors.ErrNoStateOrPKCE
	}
	// We use oatuh2.AccessTypeOffline to request a refresh token, which allows us to get a new access token when the current one expires without requiring the user to re-authenticate.
	return as.oauth.AuthCodeURL(state, oauth2.AccessTypeOffline, oauth2.S256ChallengeOption(verifier)), nil
}

// ExchangeCodeForToken exchanges the authorization code received from GitLab for an access token.
func (as *AuthService) ExchangeCodeForToken(ctx context.Context, code string, verifier string) (*oauth2.Token, error) {
	token, err := as.oauth.Exchange(ctx, code, oauth2.VerifierOption(verifier))
	if err != nil {
		return nil, err
	}
	return token, nil
}

func (as *AuthService) RefreshAccessToken(ctx context.Context, refreshToken string) (*oauth2.Token, error) {
	tokenSource := as.oauth.TokenSource(ctx, &oauth2.Token{RefreshToken: refreshToken})
	newToken, err := tokenSource.Token()
	if err != nil {
		return nil, err
	}
	return newToken, nil
}

func (as *AuthService) CreateUserWithGitlabToken(ctx context.Context, token *oauth2.Token) error {
	userData, err := as.oauth.Client(ctx, token).Get("https://gitlab.com/api/v4/user")
	if err != nil {
		return err
	}
	defer userData.Body.Close()
	userDataMap := gl.User{}
	err = json.NewDecoder(userData.Body).Decode(&userDataMap)
	if err != nil {
		return err
	}
	fmt.Printf("%v\n", userDataMap)
	return nil
}

func (as *AuthService) GetOauth2Verifier() string {
	return oauth2.GenerateVerifier()
}

func (as *AuthService) GenerateOauth2State(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	seed := rand.NewSource(time.Now().UnixNano())
	random := rand.New(seed)

	result := make([]byte, length)
	for i := range result {
		result[i] = charset[random.Intn(len(charset))]
	}
	return string(result)
}

func (as *AuthService) HandleGitlabAuthCallback(ctx context.Context, code string, verifier string) (*oauth2.Token, error) {
	token, err := as.ExchangeCodeForToken(ctx, code, verifier)
	if err != nil {
		return nil, err
	}

	err = as.CreateUserWithGitlabToken(ctx, token)
	if err != nil {
		return nil, err
	}
	return token, nil
}
