package pipelines

import (
	db "back/pkg/database"
	"errors"
	"fmt"
	"time"

	gl "github.com/xanzy/go-gitlab"
)

type Pipeline struct {
	InternalUserId int       `json:"internal_user_id,omitempty" gorm:"internal_user_id"`
	ID             int       `json:"id" gorm:"id"`
	IID            int       `json:"iid" gorm:"iid"`
	ProjectID      int       `json:"project_id" gorm:"project_id"`
	Ref            string    `json:"ref" gorm:"ref"`
	Status         string    `json:"status" gorm:"status"`
	URL            string    `json:"url" gorm:"url"`
	CreatedAt      time.Time `json:"created_at" gorm:"created_at"`
	UpdatedAt      time.Time `json:"updated_at" gorm:"updated_at"`
}

func TranslateGLPipelineToPipeline(glp *gl.Pipeline) (*Pipeline, error) {
	if glp == nil {
		return nil, errors.New("Error, invalid gitlab pipeline received")
	}

	p := Pipeline{
		ID:        glp.ID,
		IID:       glp.IID,
		ProjectID: glp.ProjectID,
		Ref:       glp.Ref,
		Status:    glp.Status,
		URL:       glp.WebURL,
		CreatedAt: *glp.CreatedAt,
		UpdatedAt: *glp.UpdatedAt,
	}
	return &p, nil
}

func init() {
	resp := db.PublicDB.AutoMigrate(&Pipeline{})
	if resp != nil {
		panic(fmt.Sprintf("Error automigrating pipelines table: %s", resp.Error()))
	}
}
