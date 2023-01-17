package keycloak

import (
	"context"
	"encoding/gob"
	"errors"
	"os"

	"github.com/Nerzal/gocloak/v12"
)

type keycloak struct {
	client       *gocloak.GoCloak
	clientID     string
	clientSecret string
	realm        string
}

var kc keycloak

func LoginUser(username string, password string) (*gocloak.JWT, error) {
	jwt, err := kc.client.Login(context.Background(), kc.clientID, kc.clientSecret, kc.realm, username, password)
	if err != nil {
		return nil, err
	}
	return jwt, nil
}

func ValidateToken(accessToken string) error {
	result, err := kc.client.RetrospectToken(context.Background(), accessToken, kc.clientID, kc.clientSecret, kc.realm)
	if err != nil {
		return err
	}
	if !*result.Active {
		return errors.New("invalid or expired token")
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

func init() {
	kc = keycloak{
		client:       gocloak.NewClient(os.Getenv("KEYCLOAK_HOST")),
		clientID:     os.Getenv("KEYCLOAK_CLIENT_ID"),
		clientSecret: os.Getenv("KEYCLOAK_SECRET"),
		realm:        os.Getenv("KEYCLOAK_REALM"),
	}
	if kc.client == nil {
		panic("Error connecting to Keycloak")
	}
	gob.Register(gocloak.JWT{})
}
