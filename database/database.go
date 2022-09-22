package database

import (
	"fmt"
	"os"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var DB *gorm.DB
var TokensDB *gorm.DB

func init() {
	db_user := os.Getenv("DB_USER")
	db_pass := os.Getenv("DB_PASS")
	db_host := os.Getenv("DB_HOST")
	db_port := os.Getenv("DB_PORT")
	db_name := os.Getenv("DB_NAME")
	tokens_db_user := os.Getenv("TOKENS_DB_USER")
	tokens_db_pass := os.Getenv("TOKENS_DB_PASS")
	tokens_db_host := os.Getenv("TOKENS_DB_HOST")
	tokens_db_port := os.Getenv("TOKENS_DB_PORT")
	tokens_db_name := os.Getenv("TOKENS_DB_NAME")
	var err error
	db_dns := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local", db_user, db_pass, db_host, db_port, db_name)
	fmt.Printf("db_dns: %s", db_dns)
	DB, err = gorm.Open(mysql.Open(db_dns), &gorm.Config{})
	if err != nil {
		panic(err.Error())
	}
	tokens_db_dns := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local", tokens_db_user, tokens_db_pass, tokens_db_host, tokens_db_port, tokens_db_name)
	fmt.Printf("tokens_db_dns: %s", tokens_db_dns)
	TokensDB, err = gorm.Open(mysql.Open(tokens_db_dns), &gorm.Config{})
	if err != nil {
		panic(err.Error())
	}
}
