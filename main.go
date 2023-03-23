package main

import (
	"back/controllers/login"
	"back/controllers/mock_data"
	"back/controllers/runners"
	"back/controllers/users"
	"back/database"
	glm "back/models/gitlab"
	"back/models/queue"
	um "back/models/users"
	"fmt"
	"math/rand"
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
		AllowOrigins:     []string{"http://localhost:8080", "http://localhost:9080", "http://localhost:3000"},
		AllowCredentials: true,
		//AllowHeaders:     []string{"Origin"},
		AllowHeaders: []string{"Content-Type", "Content-Length", "Accept-Encoding", "Authorization", "Cache-Control", "Private-Token"},
		MaxAge:       12 * time.Hour,
	}))
	authorized.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:8080", "http://localhost:9080", "http://localhost:3000"},
		AllowCredentials: true,
		//AllowHeaders:     []string{"Origin"},
		AllowHeaders: []string{"Content-Type", "Content-Length", "Accept-Encoding", "Authorization", "Cache-Control", "Private-Token"},
		MaxAge:       12 * time.Hour,
	}))
	r.POST("/login", login.ValidateLogin)
	r.POST("/register", users.Register)
	r.GET("/validate_token", users.ValidateToken)
	r.GET("/refresh_token", users.RefreshToken)
	authorized.Use(login.CheckLoginMidleware())
	authorized.Use(um.SetGitClientMiddleware())
	{
		authorized.GET("/git_runners", runners.ListUserRunnersFromGit)
		authorized.GET("/git_details", users.GetUserGitDetails)
		authorized.POST("/git_details", users.UpdateUserGitDetails)
		authorized.POST("/test_git_details", users.TestGitConnection)

		authorized.GET("/test_mt", func(c *gin.Context) {
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
		})
		authorized.GET("/get", func(c *gin.Context) {
			/* userToken := c.GetString("user_git_token")
			if userToken == "" {
				c.JSON(500, gin.H{
					"error": "No git token found in context",
				})
				return
				} */
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
	//r.GET("/confirm_registration", users.ConfirmRegistration)
	/* r.GET("/test_email_send", func(ctx *gin.Context) {
		from := "florin.rada87@yahoo.com"
		smtpApiKeyName := "apikey"
		smtpApiKey := "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"
		// Receiver email address.
		to := []string{
			"florin.rada87@yahoo.com",
		}

		// smtp server configuration.
		smtpHost := "smtp.sendgrid.net"
		smtpPort := "587"
		header := make(map[string]string)
		header["From"] = from
		header["To"] = to[0]
		header["MIME-Version"] = "1.0"
		header["Content-Type"] = "text/plain; charset=\"utf-8\""
		header["Content-Transfer-Encoding"] = "base64"
		// Message.
		message := ""
		for k, v := range header {
			message += fmt.Sprintf("%s: %s\r\n", k, v)
		}
		message += "\r\nThis is a test email message."
		fmt.Printf("the email message: %s", message)
		// Authentication.
		auth := smtp.PlainAuth("", smtpApiKeyName, smtpApiKey, smtpHost)

		// Sending email.
		err := smtp.SendMail(smtpHost+":"+smtpPort, auth, from, to, []byte(message))
		if err != nil {
			fmt.Println(err)
			return
		}
		ctx.JSON(200, gin.H{
			"error": "",
		})
	}) */
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

	for i := 0; i < 10; i++ {
		func(num int) {
			queue.AddToQueue(func(...interface{}) {
				fmt.Printf("i: %d", num)
			})
		}(i)
	}

	r.Run(":8080")
}
