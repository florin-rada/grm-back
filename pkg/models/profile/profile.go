package profile

import (
	db "back/pkg/database"
	"errors"
	"fmt"
	"time"
)

type Profile struct {
	IDUser           string `json:"id_user" gorm:"id_user,unique"`
	FirstName        string `json:"first_name" gorm:"first_name"`
	LastName         string `json:"last_name" gorm:"last_name"`
	PreferedLanguage string `json:"prefered_language" gorm:"prefered_language"`
	Country          string `json:"country" gorm:"country"`
}

type Billing struct {
	IDUser        string     `json:"id_user" gorm:"id_user,unique"`
	IDPackageType uint       `json:"id_package_type" gorm:"id_package_type"`
	StartDate     *time.Time `json:"start_date" gorm:"start_date"`
	EndDate       *time.Time `json:"end_date" gorm:"end_date"`
}

func CreateProfile(idUser string, firstName string, lastName string) (*Profile, error) {
	if idUser == "" {
		return nil, errors.New("invalid user id")
	}
	if len(firstName) < 3 {
		return nil, errors.New("invalid first name")
	}

	if len(lastName) < 3 {
		return nil, errors.New("invalid last name")
	}
	p := Profile{
		IDUser:    idUser,
		FirstName: firstName,
		LastName:  lastName,
	}
	resp := db.PublicDB.Create(&p)
	if resp.Error != nil {
		return nil, resp.Error
	}
	return &p, nil
}

func init() {
	err := db.PublicDB.AutoMigrate(&Profile{})
	if err != nil {
		panic(fmt.Sprintf("Error migrating profiles table: %s", err.Error()))
	}
	err = db.PublicDB.AutoMigrate(&Billing{})
	if err != nil {
		panic(fmt.Sprintf("Error migrating billing table: %s", err.Error()))
	}
}
