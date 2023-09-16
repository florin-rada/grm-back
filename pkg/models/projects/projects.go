package projects

import (
	"back/pkg/database"
	"fmt"

	gl "github.com/xanzy/go-gitlab"
	"gorm.io/gorm"
)

type ProjectModel struct {
	db     *gorm.DB
	client *gl.Client
}

func NewProjectModel(db *gorm.DB, client *gl.Client) *ProjectModel {
	return &ProjectModel{
		db:     db,
		client: client,
	}
}

type Project struct {
	ID             int    `gorm:"id,primarykey" json:"id"`
	InternalUserID string `gorm:"internal_user_id" json:"internal_user_id"`
	Name           string `gorm:"name" json:"name"`
}

func (pm ProjectModel) GetUserProjects(internalUserID string, page int, perPage int) ([]Project, error) {
	projects := []Project{}
	tr := pm.db.Model(&Project{})
	if page >= 0 && perPage >= 0 {
		tr = tr.Limit(perPage).Offset(perPage * page)
	}
	resp := tr.Find(&projects)
	return projects, resp.Error
}

func init() {
	resp := database.PublicDB.AutoMigrate(&Project{})
	if resp != nil {
		panic(fmt.Sprintf("Error migrating %s table: %s", "projects", resp.Error()))
	}
}
