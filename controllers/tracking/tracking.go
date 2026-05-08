package tracking

import (
	"fmt"

	trackingModel "github.com/florin-rada/grm-back/models/tracking"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type TrackingController struct {
	db *gorm.DB
	tr *trackingModel.TrackingRepository
}

func NewTrackingController(db *gorm.DB) *TrackingController {
	return &TrackingController{
		db: db,
		tr: trackingModel.NewTrackingRepository(db),
	}
}

func (tc TrackingController) AddEvent(ctx *gin.Context) {
	sess := sessions.Default(ctx)
	args := struct {
		Action   string `json:"action"`
		Target   string `json:"target"`
		Referrer string `json:"referrer"`
	}{}
	err := ctx.BindJSON(&args)
	if err != nil {
		fmt.Printf("Error, invalid arguments received for AddEvent: %s", err.Error())
		return
	}
	var UUID string
	uuidI := sess.Get("UUID")
	if uuidI == nil {
		UUID = uuid.New().String()
		sess.Set("UUID", UUID)
		sess.Save()
	} else {
		UUID = uuidI.(string)
	}
	err = tc.tr.AddEvent(UUID, trackingModel.EventType(args.Action), args.Target, args.Referrer)
	if err != nil {
		fmt.Printf("Error adding event to database: %s", err.Error())
		return
	}
}
