package authservice

import (
	"github.com/florin-rada/grm-back/config"
	"gorm.io/gorm"
)

type AuthRepository struct {
	config *config.AuthConfig
	db     *gorm.DB
}

func NewAuthRepository(db *gorm.DB, config *config.AuthConfig) *AuthRepository {
	return &AuthRepository{db: db, config: config}
}
