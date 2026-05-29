package userservice

import "github.com/florin-rada/grm-back/models/user"

type UserService struct {
	repo *UserRepository
}

func NewUserService(repo *UserRepository) *UserService {
	return &UserService{repo: repo}
}

func (us *UserService) GetUserByEmail(email string) (user.User, error) {
	u, err := us.repo.GetUserByEmail(email)
	if err != nil {
		return user.User{}, err
	}
	return u, nil
}

func (us *UserService) CreateUser(u user.User) error {
	err := us.repo.CreateUser(u)
	return err
}
