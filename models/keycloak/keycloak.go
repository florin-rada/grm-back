package keycloak

import (
	"context"
	"encoding/gob"
	"errors"
	"fmt"
	"os"

	"github.com/Nerzal/gocloak/v12"
)

type keycloak struct {
	client         *gocloak.GoCloak
	clientID       string
	clientSecret   string
	realm          string
	adminCLISecret string
}

var ErrExpiredToken = errors.New("expired token")

var kc keycloak

func LoginUser(username string, password string) (*gocloak.JWT, error) {
	jwt, err := kc.client.Login(context.Background(), kc.clientID, kc.clientSecret, kc.realm, username, password)
	if err != nil {
		return nil, err
	}
	return jwt, nil
}

func RefreshToken(refreshToken string) (*gocloak.JWT, error) {
	jwt, err := kc.client.RefreshToken(context.Background(), refreshToken, kc.clientID, kc.clientSecret, kc.realm)
	return jwt, err
}

func ValidateToken(accessToken string) error {
	result, err := kc.client.RetrospectToken(context.Background(), accessToken, kc.clientID, kc.clientSecret, kc.realm)
	if err != nil {
		return err
	}
	if !*result.Active {
		return ErrExpiredToken
	}
	return nil
}

func GetUserInfo(accessToken string) (*gocloak.UserInfo, error) {
	ui, err := kc.client.GetUserInfo(context.Background(), accessToken, kc.realm)
	if err != nil {
		return nil, err
	}
	return ui, nil
}

func CreateUser(email string, password string, firstName string, lastName string) (*string, error) {
	token, err := getAdminToken()
	if err != nil {
		return nil, err
	}
	var credentialType string = "password"
	var isTemporary bool = false
	//var isEmailVerified bool = false
	userID, err := kc.client.CreateUser(context.Background(), *token, kc.realm, gocloak.User{
		Username:  &email,
		Email:     &email,
		FirstName: &firstName,
		LastName:  &lastName,
		//EmailVerified: &isEmailVerified,
		Credentials: &[]gocloak.CredentialRepresentation{
			{
				Temporary: &isTemporary,
				Type:      &credentialType,
				Value:     &password,
			},
		},
		RequiredActions: &[]string{
			"VERIFY_EMAIL",
		},
	})
	if err != nil {
		fmt.Printf("keycloak::CreateUser error creating user: %s", err.Error())
		return nil, err
	}
	return &userID, nil
}

func getAdminToken() (*string, error) {
	jwt, err := kc.client.GetToken(context.Background(), "master", gocloak.TokenOptions{
		ClientID:     gocloak.StringP("admin-cli"),
		GrantType:    gocloak.StringP("client_credentials"),
		ClientSecret: &kc.adminCLISecret,
		//Username:     &kc.clientID,
		//Password:     &kc.clientSecret,
	})
	//jwt, err := kc.client.LoginAdmin(context.Background(), "admin", "admin", "master")
	if err != nil {
		fmt.Printf("Error getting admin token: %s", err.Error())
		return nil, err
	}
	return &jwt.AccessToken, nil
}

func ChangeUserPassword(userID string, password string) error {
	jwt, err := getAdminToken()
	if err != nil {
		return err
	}
	err = kc.client.SetPassword(context.Background(), *jwt, userID, kc.realm, password, false)
	if err != nil {
		return err
	}
	return nil
}

func init() {
	kc = keycloak{
		client:         gocloak.NewClient(os.Getenv("KEYCLOAK_HOST")),
		clientID:       os.Getenv("KEYCLOAK_CLIENT_ID"),
		clientSecret:   os.Getenv("KEYCLOAK_SECRET"),
		realm:          os.Getenv("KEYCLOAK_REALM"),
		adminCLISecret: os.Getenv("KEYCLOAK_ADMIN_CLI_SECRET"),
	}
	if kc.client == nil {
		panic("Error connecting to Keycloak")
	}
	gob.Register(gocloak.JWT{})
}
