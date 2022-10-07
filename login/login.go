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

func ValidateLogin(user string, pass string) (bool, error) {
	if user == "user" && pass == "pass" {
		return true, nil
	}
	return false, nil
}

func init() {

}
