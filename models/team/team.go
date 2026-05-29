package team

import (
	"github.com/florin-rada/grm-back/models/user"
	"gorm.io/gorm"
)

type Team struct {
	gorm.Model
	Name  string      `json:"name" gorm:"uniqueIndex:unique_team_name"`
	Users []user.User `json:"users" gorm:"many2many:team_users;"`
}
