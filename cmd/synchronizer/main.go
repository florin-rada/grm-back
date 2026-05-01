package main

import (
	"back/pkg/database"
	"back/pkg/models/gitlab"
	"back/pkg/models/jobs"
	"back/pkg/models/keycloak"
	"back/pkg/models/queue"
	"back/pkg/models/runners"
	users_model "back/pkg/models/users"
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"runtime"
	"runtime/pprof"
	"time"

	gl "gitlab.com/gitlab-org/api/client-go/v2"
	"gorm.io/gorm"
)

type UserData struct {
	GitToken      string
	BaseGitlabURL string
	MaxRunners    int
	RunnersToSync []runners.Runner
}

var userData map[string]UserData
var updateUserDataChannel chan string
var syncTycker *time.Ticker

func loadData(rr *runners.RunnerRepository) error {
	users, err := keycloak.GetUsers()
	if err != nil {
		return err
	}
	for _, u := range users {
		ud, err := getUserData(rr, *u.ID)
		if err != nil {
			return err
		}
		userData[*u.ID] = ud

	}
	return nil
}

func getUserData(rr *runners.RunnerRepository, id string) (UserData, error) {
	ud := UserData{}
	ugd, err := users_model.GetUserGitDetails(id)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return UserData{}, err
	} else if err != nil && errors.Is(err, gorm.ErrRecordNotFound) {
		ugd = &users_model.GitlabDetails{}
	}
	uo, err := users_model.GetOfferForUser(id)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return UserData{}, err
	}
	ud.MaxRunners = uo.MaxRunners
	ud.GitToken = ugd.Token
	ud.BaseGitlabURL = ugd.InstanceURL
	if ud.GitToken != "" && ud.BaseGitlabURL != "" {
		ur, err := rr.GetRunnersToSync(id, ud.MaxRunners)
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return UserData{}, err
		}
		ud.RunnersToSync = ur
	}
	return ud, nil
}

func updateUserData(rr *runners.RunnerRepository) error {
	for {
		select {
		case userID := <-updateUserDataChannel:
			{
				ud, err := getUserData(rr, userID)
				if err != nil {
					return err
				}
				userData[userID] = ud
			}
		// if we timeout and receive nothing in this time, it means there's nothing to update so we exit
		// to allow other functions to do their job
		case <-time.After(1 * time.Second):
			{
				return nil
			}
		}
	}
}

func addUpdateStatusesToQueue(qm queue.QueueManager, gitClient *gl.Client, rr *runners.RunnerRepository, runnerID uint) {
	qm.AddToQueue(func(...interface{}) {
		rr.SyncRunnerStatus(gitClient, runnerID)
	})
}

func addUpdateRunnerJobsToQueue(qm queue.QueueManager, userID string, jm *jobs.JobsModel, runnerID uint) {
	qm.AddToQueue(func(...interface{}) {
		jm.SyncRunnerJobs(userID, runnerID)
	})
}

var cpuprofile = flag.String("cpuprofile", "", "write cpu profile to file")

func main() {
	flag.Parse()
	if *cpuprofile != "" {
		fmt.Printf("\nCreating profile file at %s\n", *cpuprofile)
		f, err := os.Create(*cpuprofile)
		if err != nil {
			log.Fatal(err)
		}
		pprof.StartCPUProfile(f)
		defer pprof.StopCPUProfile()
	}
	qm := queue.NewQueueManager(runtime.NumCPU()*1000, runtime.NumCPU()*100)
	rr := runners.NewRunnerRepository(database.PublicDB)
	err := loadData(rr)
	if err != nil {
		panic(err.Error())
	}

	//counter := 0
	for {
		<-syncTycker.C

		updateUserData(rr)
		for userID, ud := range userData {
			if ud.BaseGitlabURL == "" || ud.GitToken == "" {
				continue
			}
			gitClient, err := gitlab.GetClient(ud.BaseGitlabURL, ud.GitToken)
			if err != nil {
				continue
			}
			rr := runners.NewRunnerRepository(database.PublicDB)
			jm := jobs.NewJobsModel(database.PublicDB, gitClient)
			for _, r := range ud.RunnersToSync {
				// first, update the runner statuses
				addUpdateStatusesToQueue(qm, gitClient, rr, uint(r.ID))
				// second, we update the jobs
				addUpdateRunnerJobsToQueue(qm, userID, jm, uint(r.ID))
			}
		}
	}
}

func init() {
	userData = make(map[string]UserData)
	updateUserDataChannel = make(chan string, 1000)
	syncTycker = time.NewTicker(60 * time.Second)
}
