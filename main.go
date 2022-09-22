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
