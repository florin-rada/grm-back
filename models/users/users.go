package users

import (
	db "back/database"
	"errors"
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
var EmptyEmail error = errors.New("Error, no email supplied")
var EmptyToken error = errors.New("Error, no token supplied")
var InvalidCredentials error = errors.New("Error, invalid credentials")
var AlreadyConfirmed error = errors.New("Error, already confirmed this request")

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
		return false, EmptyEmail
	}

	if password == "" {
		return false, EmptyPassword
	}

	uc, err := GetUserCredential(email)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, InvalidCredentials
		}
		return false, err
	}
	err = bcrypt.CompareHashAndPassword([]byte(uc.Password), []byte(password))
	if err != nil {
		if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
			return false, InvalidCredentials
		}
		return false, err
	}
	return true, nil
}

// CreateUser creates the new user and starts the registration confirmation process
// It also validates that the email is valid
func CreateUser(email, password string) (*User, error) {
	if email == "" {
		return nil, errors.New("Error, invalid email")
	}
	_, err := mail.ParseAddress(email)
	if err != nil {
		return nil, err
	}

	isGood, err := CheckPasswordStrength(password)
	if err != nil {
		return nil, err
	}
	if !isGood {
		return nil, PasswordMissingElements
	}

	u := User{
		Email: email,
	}
	hash, err := EncryptPassword(password)
	if err != nil {
		return nil, err
	}
	err = db.DB.Transaction(func(tx *gorm.DB) error {
		resp := tx.Create(&u)
		if resp.Error != nil {
			return resp.Error
		}

		uc := UserCredential{
			IDUser:   u.ID,
			Password: hash,
		}

		resp = tx.Create(&uc)
		if err != nil {
			return resp.Error
		}

		err = BeginRegistrationFlow(tx, &u)
		if err != nil {
			return err
		}
		return nil
	})
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

func GetUser(email string) (*User, error) {
	if email == "" {
		return nil, EmptyEmail
	}
	u := User{}
	resp := db.DB.Model(&u).Where("email=?", email).Find(&u)
	if resp.Error != nil {
		return nil, resp.Error
	}
	return &u, nil
}
func GetUserCredential(email string) (UserCredential, error) {
	uc := UserCredential{}
	resp := db.TokensDB.Where("email=?", email).Find(&uc)
	if resp.Error != nil {
		return UserCredential{}, resp.Error
	}
	return uc, nil
}

func EncryptPassword(password string) (string, error) {
	if password == "" {
		return "", EmptyEmail
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(password), cost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

func BeginRegistrationFlow(tx *gorm.DB, u *User) error {
	newUUID := uuid.New()
	regDate := time.Now()
	r := Registration{
		IDUser:           u.ID,
		Token:            newUUID.String(),
		Confirmed:        false,
		RegistrationDate: &regDate,
	}
	resp := tx.Create(&r)
	if resp.Error != nil {
		return resp.Error
	}

	// send email

	return nil
}

func ConfirmRegistration(tx *gorm.DB, token string) (bool, error) {
	if token == "" {
		return false, errors.New("Error, no token supplied")
	}

	r := Registration{}
	resp := db.TokensDB.Model(&r).Where("token=?", token).Find(&r)
	if resp.Error != nil {
		return false, resp.Error
	}
	if r.ConfirmationDate != nil {
		return false, AlreadyConfirmed
	}

	if r.Confirmed {
		return false, AlreadyConfirmed
	}
	cd := time.Now()
	r.ConfirmationDate = &cd
	r.Confirmed = true
	resp = tx.Save(&r)
	if resp.Error != nil {
		return false, resp.Error
	}
	return true, nil
}

func GetRegistrationForUser(id uint) (*Registration, error) {
	r := Registration{}
	resp := db.TokensDB.Find(&r, id)
	if resp.Error != nil {
		return nil, resp.Error
	}
	return &r, nil
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
