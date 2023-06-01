package mock_data

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

type User struct {
	ID        uint
	FirstName string
	LastName  string
	Email     string
	Gender    string
	IpAddress string
}

var users []User

func ReturnMockData(ctx *gin.Context) {
	searchTerm := ctx.Query("search_term")
	page := 1
	perPage := 25
	pageStr := ctx.Query("page")
	if pageStr != "" {
		tmpPage, err := strconv.ParseInt(pageStr, 10, 0)
		if err != nil {
			ctx.JSON(500, gin.H{
				"error":    err.Error(),
				"response": "",
				"total":    "",
			})
			return
		}
		page = int(tmpPage)
	}
	perPageStr := ctx.Query("per_page")
	if perPageStr != "" {
		tmpPerPage, err := strconv.ParseInt(perPageStr, 10, 0)
		if err != nil {
			ctx.JSON(500, gin.H{
				"error":    err.Error(),
				"response": "",
				"total":    "",
			})
			return
		}
		perPage = int(tmpPerPage)
	}
	fmt.Printf("page: %s, perPage: %s", pageStr, perPageStr)
	// search for value
	if searchTerm != "" {
		toReturn := []User{}
		for _, u := range users {
			if strings.Contains(u.FirstName, searchTerm) {
				toReturn = append(toReturn, u)
				continue
			}
			if strings.Contains(u.LastName, searchTerm) {
				toReturn = append(toReturn, u)
				continue
			}
			if strings.Contains(u.Email, searchTerm) {
				toReturn = append(toReturn, u)
				continue
			}
			if strings.Contains(u.IpAddress, searchTerm) {
				toReturn = append(toReturn, u)
				continue
			}
			if strings.Contains(u.Gender, searchTerm) {
				toReturn = append(toReturn, u)
				continue
			}
			if strings.Contains(fmt.Sprintf("%d", u.ID), searchTerm) {
				toReturn = append(toReturn, u)
				continue
			}
		}
		ctx.JSON(200, gin.H{
			"error":    "",
			"response": toReturn,
			"total":    len(toReturn),
		})
		return
	} else {
		ctx.JSON(200, gin.H{
			"error":    "",
			"response": users[(page-1)*perPage : ((page-1)*perPage)+perPage],
			"total":    len(users),
		})
	}
}

func init() {
	data, err := os.ReadFile("MOCK_DATA.csv")
	if err != nil {
		panic(fmt.Sprintf("Error reading mock data: %s", err.Error()))
	}
	rows := strings.Split(string(data), "\r\n")
	for lineNo, row := range rows {
		if lineNo == 0 || row == "" {
			continue
		}
		cols := strings.Split(row, ",")
		//fmt.Printf("cols: %+v", cols)
		//fmt.Printf(" id: %s\n", cols[0])
		id, err := strconv.ParseInt(cols[0], 10, 0)
		if err != nil {
			panic(fmt.Sprintf("Error parsing id: %s", err.Error()))
		}
		u := User{
			ID:        uint(id),
			FirstName: cols[1],
			LastName:  cols[2],
			Email:     cols[3],
			Gender:    cols[4],
			IpAddress: cols[5],
		}
		users = append(users, u)
	}
}
