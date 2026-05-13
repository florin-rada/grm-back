package team

import (
	"gorm.io/gorm"
)

type Team struct {
	gorm.Model
	Name string `json:"name" gorm:"uniqueIndex:unique_team_name"`
}
