package login

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

func ValidateLogin(userEmail string, pass string) (bool, error) {
	if userEmail == "user" && pass == "pass" {
		return true, nil
	}

	//_, _ := users.ValidateUserPassword(userEmail, pass)
	return true, nil
}

func init() {

}
