package pipelines

import (
	"fmt"
	"time"

	db "github.com/florin-rada/grm-back/database"
)

type Pipeline struct {
	InternalUserId int64     `json:"internal_user_id,omitempty" gorm:"internal_user_id"`
	ID             int64     `json:"id" gorm:"id"`
	IID            int64     `json:"iid" gorm:"iid"`
	ProjectID      int64     `json:"project_id" gorm:"project_id"`
	Ref            string    `json:"ref" gorm:"ref"`
	Status         string    `json:"status" gorm:"status"`
	URL            string    `json:"url" gorm:"url"`
	CreatedAt      time.Time `json:"created_at" gorm:"created_at"`
	UpdatedAt      time.Time `json:"updated_at" gorm:"updated_at"`
}

func init() {
	resp := db.PublicDB.AutoMigrate(&Pipeline{})
	if resp != nil {
		panic(fmt.Sprintf("Error automigrating pipelines table: %s", resp.Error()))
	}
}
