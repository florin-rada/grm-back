package userservice

import (
	"github.com/florin-rada/grm-back/models/user"

	"gorm.io/gorm"
)

type UserRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (ur UserRepository) GetUserByEmail(email string) (user.User, error) {
	var u user.User
	result := ur.db.Where("email = ?", email).First(&u)
	if result.Error != nil {
		return user.User{}, result.Error
	}
	return u, nil
}

func (ur UserRepository) CreateUser(u user.User) error {
	result := ur.db.Create(&u)
	return result.Error
}
