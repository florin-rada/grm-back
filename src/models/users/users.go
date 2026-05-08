package users

import (
	"errors"
	"fmt"
	"net/http"
	"net/mail"
	re "regexp"
	"time"

	consterrors "github.com/florin-rada/grm-back/src/const_errors"
	db "github.com/florin-rada/grm-back/src/database"
	"github.com/florin-rada/grm-back/src/models/gitlab"
	"github.com/florin-rada/grm-back/src/models/keycloak"
	"github.com/florin-rada/grm-back/src/models/offers"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

const cost = 12

var ErrEmptyPassword error = errors.New("empty password")
var ErrPasswordToShort error = errors.New("password shorter than 10 characters")
var ErrPasswordMissingElements error = errors.New("password must containt lower case, upper case, digit and at least a symbol")
var ErrEmptyEmail error = errors.New("no email supplied")
var ErrEmptyToken error = errors.New("no token supplied")
var ErrInvalidCredentials error = errors.New("invalid credentials")
var ErrAlreadyConfirmed error = errors.New("already confirmed this request")

type PublicModel struct {
	PublicDB *gorm.DB
}

type PrivateModel struct {
	PrivateDB *gorm.DB
}

type User struct {
	gorm.Model
	Email string `json:"email" gorm:"email,uniqueIndex:unique_email"`
}

type UserOffer struct {
	IDUser  string       `gorm:"id_user,uniqueIndex:unique_offer"`
	IDOffer int          `gorm:"id_offer"`
	Offer   offers.Offer `gorm:"foreignKey:IDOffer"`
}

type UserCredential struct {
	IDUser   uint   `json:"id_user" gorm:"uniqueIndex;not null"`
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
	IDUser           uint       `json:"id_user" gorm:"id_user;uniqueIndex;not null"`
	Token            string     `json:"token" gorm:"token"`
	RequestConfirmed bool       `json:"request_confirmed" gorm:"request_confirmed"`
	StartTime        *time.Time `json:"start_time" gorm:"start_time"`
	ConfirmationTime *time.Time `json:"confirmation_time" gorm:"confirmation_time"`
	ToDeleteTime     *time.Time `json:"to_delete_time" gorm:"to_delete_time"`
}

type Registration struct {
	IDUser           uint       `json:"id_user" gorm:"id_user;primarykey;uniqueIndex;not null"`
	Token            string     `json:"token" gorm:"token"`
	Confirmed        bool       `json:"confirmed" gorm:"confirmed"`
	RegistrationDate *time.Time `json:"registration_date" gorm:"registration_date"`
	ConfirmationDate *time.Time `json:"confirmation_date" gorm:"confirmation_date"`
}

type PasswordReset struct {
	IDUser      uint       `json:"id_user" gorm:"id_user;primarykey;unique"`
	Token       string     `json:"token" gorm:"token"`
	RequestDate *time.Time `json:"request_date" gorm:"request_date"`
	BestBefore  *time.Time `json:"best_before" gorm:"best_before"`
	Confirmed   bool       `json:"confirmed" gorm:"confirmed"`
}

type GitlabDetails struct {
	IDUser      string `json:"id_user" gorm:"id_user;unique"`
	Token       string `json:"token" gorm:"token"` // it must NEVER be returned or exposed by controllers
	InstanceURL string `json:"instance_url" gorm:"instance_url"`
	SyncRate    int    `json:"sync_rate" gorm:"sync_rate"`
}

func ValidateUserPassword(email, password string) (bool, error) {
	if email == "" {
		return false, ErrEmptyEmail
	}

	if password == "" {
		return false, ErrEmptyPassword
	}

	uc, err := GetUserCredential(email)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, ErrInvalidCredentials
		}
		return false, err
	}
	err = bcrypt.CompareHashAndPassword([]byte(uc.Password), []byte(password))
	if err != nil {
		if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
			return false, ErrInvalidCredentials
		}
		return false, err
	}
	return true, nil
}

// CreateUser creates the new user and starts the registration confirmation process
// It also validates that the email is valid
func CreateUser(email, password string, firstName string, lastName string) (string, error) {
	if email == "" {
		return "", errors.New("invalid email")
	}
	_, err := mail.ParseAddress(email)
	if err != nil {
		return "", err
	}

	isGood, err := CheckPasswordStrength(password)
	if err != nil {
		return "", err
	}
	if !isGood {
		return "", ErrPasswordMissingElements
	}
	// we create the keycloak user
	kcUserId, err := keycloak.CreateUser(email, password, firstName, lastName)
	if err != nil {
		return "", err
	}

	// create the gitlab details instance, will hold the user's api endpoint and gitlab token
	gitDetails := GitlabDetails{
		*kcUserId,
		"",
		"",
		0,
	}
	resp := db.PrivateDB.Save(&gitDetails)
	if resp.Error != nil {
		return "", err
	}
	fmt.Printf("New user id for : %s : %s", email, *kcUserId)
	//p, err := profile.CreateProfile(*kcUserId, )
	/* hash, err := EncryptPassword(password)
	if err != nil {
		return nil, err
	}
	err = db.PublicDB.Transaction(func(tx *gorm.DB) error {
		resp := tx.Create(&u)
		if resp.Error != nil {
			return resp.Error
		}

		uc := UserCredential{
			IDUser:   u.ID,
			Password: hash,
		}
		err := db.PrivateDB.Transaction(func(tx *gorm.DB) error {
			resp = tx.Create(&uc)
			if resp.Error != nil {
				return resp.Error
			}

			err = BeginRegistrationFlow(tx, &u)
			return err
		})

		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return nil, err
	} */
	return *kcUserId, nil
}

func GetUserGitDetails(idUser string) (*GitlabDetails, error) {
	gld := GitlabDetails{}
	resp := db.PrivateDB.Model(GitlabDetails{}).Where("id_user=?", idUser).First(&gld)
	if resp.Error != nil {
		return nil, resp.Error
	}
	return &gld, nil
}

func SetUserGitDetails(gld *GitlabDetails) error {
	if gld == nil {
		return consterrors.ErrNoGitToken
	}
	if gld.IDUser == "" {
		return consterrors.ErrNoUserId
	}
	resp := db.PrivateDB.Save(&gld)
	if resp.Error != nil {
		return resp.Error
	}
	return nil
}

func GetOfferForUser(idUser string) (offers.Offer, error) {
	var uo UserOffer
	resp := db.PublicDB.Model(&UserOffer{}).Where("id_user=?", idUser).First(&uo)
	if errors.Is(resp.Error, gorm.ErrRecordNotFound) {
		or := offers.NewOfferRepository(db.PublicDB)
		defaultOffer, err := or.GetDefaultOffer()
		return defaultOffer, err
	}
	if resp.Error != nil {
		return offers.Offer{}, resp.Error
	}
	return uo.Offer, nil
}

func SetGitClientMiddleware() gin.HandlerFunc {
	return func(next *gin.Context) {
		userID, exists := next.Get("user_id")
		if !exists {
			next.Next()
			return
		}
		gitDetails, err := GetUserGitDetails(userID.(string))
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			next.AbortWithError(http.StatusInternalServerError, err)
			return
		} else if err != nil && errors.Is(err, gorm.ErrRecordNotFound) {
			next.Next()
			return
		}
		clt, err := gitlab.GetClient(gitDetails.InstanceURL, gitDetails.Token)
		if err != nil {
			next.AbortWithError(http.StatusInternalServerError, err)
			return
		}

		next.Set("git_client", clt)
		next.Next()
	}
}

func CheckPasswordStrength(password string) (bool, error) {
	if password == "" {
		return false, ErrEmptyPassword
	}
	if len(password) < 10 {
		return false, ErrPasswordToShort
	}
	match, err := re.Match(`[0-9]*`, []byte(password))
	if err != nil {
		return false, errors.New("Error, checking password strength")
	}
	if !match {
		return false, ErrPasswordMissingElements
	}
	match, err = re.Match(`[a-zA-Z]*`, []byte(password))
	if err != nil {
		return false, errors.New("Error, checking password strength")
	}
	if !match {
		return false, ErrPasswordMissingElements
	}
	match, err = re.Match(`[!@#%^&*()_\-=+,.\/?~\{\}:;'"\\|<>]*`, []byte(password))
	if err != nil {
		return false, errors.New("Error, checking password strength")
	}
	if !match {
		return false, ErrPasswordMissingElements
	}
	return true, nil
}

func GetUser(email string) (*User, error) {
	if email == "" {
		return nil, ErrEmptyEmail
	}
	u := User{}
	resp := db.PublicDB.Model(&u).Where("email=?", email).First(&u)
	if resp.Error != nil {
		return nil, resp.Error
	}
	return &u, nil
}

func GetUserCredential(email string) (UserCredential, error) {
	uc := UserCredential{}
	resp := db.PrivateDB.Where("email=?", email).Find(&uc)
	if resp.Error != nil {
		return UserCredential{}, resp.Error
	}
	return uc, nil
}

func EncryptPassword(password string) (string, error) {
	if password == "" {
		return "", ErrEmptyEmail
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

func ConfirmRegistration(tx *gorm.DB, token string) error {
	if token == "" {
		return ErrEmptyToken
	}

	r := Registration{}
	resp := tx.Model(&r).Where("token=?", token).Find(&r)
	if resp.Error != nil {
		return resp.Error
	}
	if r.ConfirmationDate != nil {
		return ErrAlreadyConfirmed
	}

	if r.Confirmed {
		return ErrAlreadyConfirmed
	}
	cd := time.Now()
	r.ConfirmationDate = &cd
	r.Confirmed = true
	resp = tx.Where("id_user=?", r.IDUser).Save(&r)
	if resp.Error != nil {
		return resp.Error
	}
	return nil
}

func GetRegistrationForUser(id uint) (*Registration, error) {
	r := Registration{}
	resp := db.PrivateDB.Find(&r, id)
	if resp.Error != nil {
		return nil, resp.Error
	}
	return &r, nil
}

func GetUnconfirmedRegistrations(days_passed uint) ([]*Registration, error) {
	registrations := []*Registration{}
	resp := db.PrivateDB.Where("`confirmed`=0 and `registration_date` < NOW() - INTERVAL ? DAY", days_passed).Find(&registrations)
	if resp.Error != nil {
		return nil, resp.Error
	}
	return registrations, nil
}

func init() {
	err := db.PublicDB.AutoMigrate(&User{})
	if err != nil {
		panic("Error migrating users table")
	}
	err = db.PublicDB.AutoMigrate(&UserOffer{})
	if err != nil {
		panic("Error migrating user_offer table")
	}
	err = db.PrivateDB.AutoMigrate(&UserCredential{})
	if err != nil {
		panic("Error migrating user_credentials table")
	}
	err = db.PrivateDB.AutoMigrate(&UserToken{})
	if err != nil {
		panic("Error migrating user_tokens table")
	}
	err = db.PrivateDB.AutoMigrate(&GitlabDetails{})
	if err != nil {
		panic("Error migrating table gitlab_tokens table")
	}

	err = db.PrivateDB.AutoMigrate(&Registration{})
	if err != nil {
		panic("Error migrating registrations table")
	}

	err = db.PrivateDB.AutoMigrate(&RequestDeletion{})
	if err != nil {
		panic("Error migrating request_deletions table")
	}

}
