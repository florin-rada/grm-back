package users

import (
	db "back/database"
	"errors"
	"fmt"
	"net/mail"
	re "regexp"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

const cost = 12

var EmptyPassword error = errors.New("Error, empty password")
var PasswordToShort error = errors.New("Error, password shorter than 10 characters")
var PasswordMissingElements error = errors.New("Error, password must containt lower case, upper case, digit and at least a symbol")

type User struct {
	gorm.Model
	Email string `json:"email" gorm:"email,unique"`
}

type UserCredential struct {
	IDUser   uint   `json:"id_user" gorm:"id_user,unique"`
	Password string `json:"password" gorm:"password"`
}

type UserToken struct {
	IDUser             uint       `json:"id_user" gorm:"id_user"`
	Token              string     `json:"token" gorm:"token"`
	RefreshToken       string     `json:"refresh_token" gorm:"refresh_token"`
	DeviceID           string     `json:"device_id" gorm:"device_id"`
	LoginDate          *time.Time `json:"login_date" gorm:"login_date"`
	ExpireAfter        *time.Time `json:"expire_after" gorm:"expire_after"`
	RefreshExpireAfter *time.Time `json:"refresh_expire_after" gorm:"refresh_expire_after"`
}

type RequestDeletion struct {
	IDUser           uint       `json:"id_user" gorm:"id_user,unique"`
	Token            string     `json:"token" gorm:"token"`
	RequestConfirmed bool       `json:"request_confirmed" gorm:"request_confirmed"`
	StartTime        *time.Time `json:"start_time" gorm:"start_time"`
	ConfirmationTime *time.Time `json:"confirmation_time" gorm:"confirmation_time"`
	ToDeleteTime     *time.Time `json:"to_delete_time" gorm:"to_delete_time"`
}

type Registration struct {
	IDUser           uint       `json:"id_user" gorm:"id_user,unique"`
	Token            string     `json:"token" gorm:"token"`
	Confirmed        bool       `json:"confirmed" gorm:"confirmed"`
	RegistrationDate *time.Time `json:"registration_date" gorm:"registration_date"`
	ConfirmationDate *time.Time `json:"confirmation_date" gorm:"confirmation_date"`
}

type PasswordReset struct {
	IDUser      uint       `json:"id_user" gorm:"id_user,unique"`
	Token       string     `json:"token" gorm:"token"`
	RequestDate *time.Time `json:"request_date" gorm:"request_date"`
	BestBefore  *time.Time `json:"best_before" gorm:"best_before"`
	Confirmed   bool       `json:"confirmed" gorm:"confirmed"`
}

type GitlabToken struct {
	IDUser uint   `json:"id_user" gorm:"id_user,unique"`
	Token  string `json:"token" gorm:"token"`
}

func ValidateUserPassword(email, password string) (bool, error) {
	if email == "" {
		return false, errors.New("Error, no email supplied")
	}

	if password == "" {
		return false, errors.New("Error, no password supplied")
	}

	uc, err := GetUserCredential(email)
	if err != nil {
		return false, err
	}
	err = bcrypt.CompareHashAndPassword([]byte(uc.Password), []byte(password))
	if err != nil {
		return false, err
	}
	return true, nil
}

func RegisterUser(email, password string) (User, error) {
	if email == "" {
		return User{}, errors.New("Error, invalid email")
	}
	_, err := mail.ParseAddress(email)
	if err != nil {
		return User{}, err
	}

	u := User{}
	return u, nil
}

func CheckPasswordStrength(password string) (bool, error) {
	if password == "" {
		return false, EmptyPassword
	}
	if len(password) < 10 {
		return false, PasswordToShort
	}
	match, err := re.Match(`[0-9]*`, []byte(password))
	if err != nil {
		return false, errors.New("Error, checking password strength")
	}
	if !match {
		return false, PasswordMissingElements
	}
	match, err = re.Match(`[a-zA-Z]*`, []byte(password))
	if err != nil {
		return false, errors.New("Error, checking password strength")
	}
	if !match {
		return false, PasswordMissingElements
	}
	match, err = re.Match(`[!@#%^&*()_\-=+,.\/?~\{\}:;'"\\|<>]*`, []byte(password))
	if err != nil {
		return false, errors.New("Error, checking password strength")
	}
	if !match {
		return false, PasswordMissingElements
	}
	return true, nil
}

func GetUserCredential(email string) (UserCredential, error) {
	uc := UserCredential{}
	resp := db.TokensDB.Where("email=?", email).Find(&uc)
	if resp.Error != nil {
		return UserCredential{}, resp.Error
	}
	return uc, nil
}

func BeginRegistrationFlow(tx *gorm.DB, email string) bool {
	newUUID := uuid.New()
	fmt.Printf("newUUID: %s", newUUID.String())
	return true
}

func init() {
	err := db.DB.AutoMigrate(&User{})
	if err != nil {
		panic("Error migrating users table")
	}
	err = db.TokensDB.AutoMigrate(&UserCredential{})
	if err != nil {
		panic("Error migrating user_credentials table")
	}
	err = db.TokensDB.AutoMigrate(&UserToken{})
	if err != nil {
		panic("Error migrating user_tokens table")
	}
	err = db.TokensDB.AutoMigrate(&GitlabToken{})
	if err != nil {
		panic("Error migrating table gitlab_tokens table")
	}

	err = db.TokensDB.AutoMigrate(&Registration{})
	if err != nil {
		panic("Error migrating registrations table")
	}

	err = db.TokensDB.AutoMigrate(&RequestDeletion{})
	if err != nil {
		panic("Error migrating request_deletions table")
	}
}
