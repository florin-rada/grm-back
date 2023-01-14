package login

import (
	"back/models/keycloak"

	"github.com/Nerzal/gocloak/v12"
)

/* type AuthData struct {
	Username string `header:"username"`
	Password string `header:"password"`
} */

type PrivateTokenStruct struct {
	PrivateToken string `header:"private_token"`
}

func ValidateToken(token string) (bool, error) {
	return true, nil
}

func ValidateLogin(userEmail string, pass string) (*gocloak.JWT, error) {
	jwt, err := keycloak.LoginUser(userEmail, pass)
	if err != nil {
		return nil, err
	}
	return jwt, nil
}

func init() {

}
