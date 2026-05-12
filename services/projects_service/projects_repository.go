package projectsservice

import "gorm.io/gorm"

type ProjectsRepository struct {
	db *gorm.DB
}

func NewProjectsRepository(db *gorm.DB) *ProjectsRepository {
	return &ProjectsRepository{db: db}
}
