package main

import (
	"back/pkg/controllers/login"
	"back/pkg/controllers/mock_data"
	"back/pkg/controllers/offers"
	"back/pkg/controllers/runners"
	"back/pkg/controllers/tracking"
	"back/pkg/controllers/users"
	"back/pkg/database"
	glm "back/pkg/models/gitlab"
	um "back/pkg/models/users"
	"os"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/sessions"
	"github.com/xanzy/go-gitlab"

	//"github.com/gin-contrib/sessions/cookie"
	gormsessions "github.com/gin-contrib/sessions/gorm"
	"github.com/gin-gonic/gin"
)

/* func testingMT() {
	client := gitlab.GetClient()
	runners, _, err := gitlab.GetAllRunners(client, 1, 10)
	if err != nil {
		fmt.Printf("Error getting all runners: %s", err.Error())
		return
	}
	output := make(chan struct {
		Jobs []*gl.Job
		Err  error
	}, len(runners))
	defer close(output)
	for _, r := range runners {
		action := func() {
			j, _, err := gitlab.GetRunnerJobs(client, r.ID, "", 1, 10)
			output <- struct {
				Jobs []*gl.Job
				Err  error
			}{Jobs: j, Err: err}
		}
		jqi := jobs.JobsQueueItem{
			Task: action,
		}
		queue.AddToQueue(jqi)
	}

	for _, _ = range runners {
		resp := <-output
		fmt.Printf("Received output : %+v\n", resp)
	}
} */

func main() {

	r := gin.Default()
	store := gormsessions.NewStore(database.PrivateDB, true, []byte(os.Getenv("SESSION_SECRET")))
	store.Options(sessions.Options{
		MaxAge:   60 * 60 * 24,
		Secure:   false,
		HttpOnly: true,
	})
	r.Use(sessions.Sessions("GILMO_SESSION", store))
	authorized := r.Group("authorized")
	authorized.Use(sessions.Sessions("GILMO_SESSION", store))
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:8080", "http://localhost:9080", "http://localhost:3000", "http://localhost:5173"},
		AllowCredentials: true,
		AllowMethods:     []string{"GET", "PUT", "POST", "DELETE", "PATCH", "OPTIONS", "HEAD"},
		//AllowHeaders:     []string{"Origin"},
		AllowHeaders:  []string{"Content-Type", "Content-Length", "Accept-Encoding", "Authorization", "Cache-Control", "Private-Token"},
		MaxAge:        12 * time.Hour,
		AllowWildcard: true,
	}))
	authorized.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:8080", "http://localhost:9080", "http://localhost:3000", "http://localhost:5173"},
		AllowCredentials: true,
		AllowMethods:     []string{"GET", "PUT", "POST", "DELETE", "PATCH", "OPTIONS", "HEAD"},
		//AllowHeaders:     []string{"Origin"},
		AllowHeaders:  []string{"Content-Type", "Content-Length", "Accept-Encoding", "Authorization", "Cache-Control", "Private-Token"},
		MaxAge:        12 * time.Hour,
		AllowWildcard: true,
	}))

	oc := offers.NewOfferController(database.PublicDB)
	rc := runners.NewRunnerController(database.PublicDB)
	tc := tracking.NewTrackingController(database.PublicDB)
	r.POST("/login", login.ValidateLogin)
	r.POST("/register", users.Register)
	r.GET("/validate_token", users.ValidateToken)
	r.GET("/refresh_token", users.RefreshToken)
	r.POST("/tracking", tc.AddEvent)
	authorized.Use(login.CheckLoginMidleware())
	authorized.Use(um.SetGitClientMiddleware())
	{
		authorized.GET("/git_runners", rc.ListUserRunnersFromGit)
		authorized.GET("/git_details", users.GetUserGitDetails)
		authorized.POST("/git_details", users.UpdateUserGitDetails)
		authorized.POST("/test_git_details", users.TestGitConnection)
		authorized.GET("/test_get_jobs", rc.TestGetJobsBetween)
		authorized.GET("/runners", rc.ListUserRunners)
		authorized.GET("/runners/from_gitlab", rc.ListUserRunnersFromGit)
		authorized.GET("/runners/:id", rc.GetRunnerDetails)
		authorized.PUT("/runners/:id", rc.UpdateRunner)
		authorized.DELETE("/runners/:id", rc.DeleteRunner)
		authorized.GET("/runners/:id/latest_jobs", rc.GetLatestJobsForRunner)
		authorized.GET("/runners/:id/jobs", rc.ListRunnerJobs)
		authorized.POST("/runners/:id", rc.AddRunnerForUser)
		authorized.GET("/offers", oc.GetOffers)
		authorized.POST("/offers", oc.CreateOffer)
		authorized.PUT("/offers/:id", oc.UpdateOffer)
		authorized.DELETE("/offers/:id", oc.DeleteOffer)
		authorized.GET("/offers/:id", oc.GetOffer)
		/* authorized.GET("/test_mt", func(c *gin.Context) {
			replyChan := make(chan int, 3)
			min := int(0)
			max := int(100)
			numToGen := 10
			func(out chan int, min int, max int, toGen int) {
				queue.AddToQueue(func(...interface{}) {
					for i := 0; i < numToGen; i++ {
						num := rand.Intn(max)
						replyChan <- num
					}
					close(replyChan)
				})
			}(replyChan, min, max, numToGen)
			generatedNum := []int{}
			for i := range replyChan {
				generatedNum = append(generatedNum, i)
			}
			c.JSON(200, gin.H{
				"generated_numbers": generatedNum,
			})
		}) */
		authorized.GET("/get", func(c *gin.Context) {
			userID, exists := c.Get("user_id")
			if !exists {
				c.JSON(500, gin.H{
					"error": "user_info not found",
				})
				return
			}
			userEmail, exists := c.Get("email")
			if !exists {
				c.JSON(500, gin.H{
					"error": "user email not found",
				})
				return
			}
			username, exists := c.Get("username")
			if !exists {
				c.JSON(500, gin.H{
					"error": "username not found",
				})
				return
			}

			gitClientI, exists := c.Get("git_client")
			if !exists {
				c.JSON(500, gin.H{
					"error": "git client not found in context",
				})
				return
			}
			gitClient := gitClientI.(*gitlab.Client)

			jobs, _, err := glm.GetAllRunners(gitClient, 1, 20)
			if err != nil {
				c.JSON(500, gin.H{
					"error": "error getting users runners",
				})
				return
			}

			c.JSON(200, gin.H{
				//"OK":            userToken,
				"userID":        userID,
				"GitlabDetails": gitClient,
				"Email":         userEmail,
				"username":      username,
				"jobs":          jobs,
			})

		})
	}

	r.GET("/mock_data", mock_data.ReturnMockData)
	r.StaticFile("/", "../../front/build/index.html")
	r.Static("/static", "../../front/build/static")

	r.StaticFile("/logo512.png", "../../front/build/logo512.png")
	r.StaticFile("/logo192.png", "../../front/build/logo192.png")
	r.StaticFile("/favicon.ico", "../../front/build/favicon.ico")
	r.StaticFile("/robots.txt.ico", "../../front/build/robots.txt.ico")
	r.StaticFile("/asset-manifest.json", "../../front/build/asset-manifest.json")
	r.StaticFile("/manifest.json", "../../front/build/manifest.json")
	//fmt.Printf("hello, world\n")
	/* res := database.PublicDB.Raw("SHOW TABLES")
	if res.Error != nil {
		panic(res.Error.Error())
	}
	res = database.PrivateDB.Raw("SHOW TABLES")
	if res.Error != nil {
		panic(res.Error.Error())
	} */

	/* for i := 0; i < 10; i++ {
		func(num int) {
			queue.AddToQueue(func(...interface{}) {
				fmt.Printf("i: %d", num)
			})
		}(i)
	} */

	r.Run(":8080")
}
