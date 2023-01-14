package main

import (
	"back/controllers/login"
	"back/controllers/mock_data"
	"back/controllers/users"
	"back/database"
	"time"

	"github.com/gin-contrib/cors"

	//"github.com/gin-contrib/sessions/cookie"

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
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:8080", "http://localhost:9080", "http://localhost:3000"},
		AllowCredentials: true,
		//AllowHeaders:     []string{"Origin"},
		AllowHeaders: []string{"Content-Type", "Content-Length", "Accept-Encoding", "Authorization", "Cache-Control", "Private-Token"},
		MaxAge:       12 * time.Hour,
	}))
	authorized := r.Group("authorized")
	/* store := gormsessions.NewStore(database.PrivateDB, true, []byte(os.Getenv("SESSION_SECRET")))
	store.Options(sessions.Options{
		MaxAge:   60 * 60 * 24,
		Secure:   false,
		HttpOnly: false,
	}) */
	//r.Use(sessions.Sessions("GILMO_SESSION", store))
	//authorized.Use(sessions.Sessions("GILMO_SESSION", store))
	authorized.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:8080", "http://localhost:9080", "http://localhost:3000"},
		AllowCredentials: true,
		//AllowHeaders:     []string{"Origin"},
		AllowHeaders: []string{"Content-Type", "Content-Length", "Accept-Encoding", "Authorization", "Cache-Control", "Private-Token"},
		MaxAge:       12 * time.Hour,
	}))
	authorized.Use(login.CheckLoginMidleware())
	{
		authorized.GET("/get", func(c *gin.Context) {
			c.JSON(200, gin.H{
				"OK": "OK",
			})
		})
	}
	r.POST("/login", login.ValidateLogin)
	r.POST("/register", users.Register)
	r.GET("/confirm_registration", users.ConfirmRegistration)

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
	res := database.PublicDB.Raw("SHOW TABLES")
	if res.Error != nil {
		panic(res.Error.Error())
	}
	res = database.PrivateDB.Raw("SHOW TABLES")
	if res.Error != nil {
		panic(res.Error.Error())
	}

	r.Run(":8080")
}
