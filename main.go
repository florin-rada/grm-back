package main

import (
	"back/controllers/mock_data"
	"back/database"
	"back/login"
	"fmt"

	"github.com/gin-gonic/gin"
)

func main() {

	r := gin.Default()
	authorized := r.Group("/v1")
	authorized.Use(login.CheckLoginMidleware())
	{
		authorized.GET("get", func(c *gin.Context) {
			c.JSON(200, gin.H{
				"OK": "OK",
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
	fmt.Printf("hello, world\n")
	res := database.DB.Raw("SHOW TABLES")
	if res.Error != nil {
		panic(res.Error.Error())
	}
	res = database.TokensDB.Raw("SHOW TABLES")
	if res.Error != nil {
		panic(res.Error.Error())
	}
	r.Run(":8080")
}
