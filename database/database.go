package database

import (
	"context"
	"database/sql"
	"fmt"
	"os"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type DBInterface interface {
	AddError(err error) error
	Assign(attrs ...interface{}) *DBInterface
	Association(column string) *gorm.Association
	Attrs(attrs ...interface{}) *DBInterface
	AutoMigrate(dst ...interface{}) error
	Begin(opts ...*sql.TxOptions) *DBInterface
	Clauses(conds ...clause.Expression) *DBInterface
	Commit() *DBInterface
	Connection(fc func(tx *DBInterface) error) error
	Count(count *int64) *DBInterface
	Create(value interface{}) *DBInterface
	CreateInBatches(value interface{}, batchSize int) *DBInterface
	DB() (*sql.DB, error)
	Debug() *DBInterface
	Delete(value interface{}, conds ...interface{}) *DBInterface
	Distinct(args ...interface{}) *DBInterface
	Exec(sql string, values ...interface{}) *DBInterface
	Find(dest interface{}, conds ...interface{}) *DBInterface
	FindInBatches(dest interface{}, batchSize int, fc func(tx *DBInterface, batch int) error) *DBInterface
	First(dest interface{}, conds ...interface{}) *DBInterface
	FirstOrCreate(dest interface{}, conds ...interface{}) *DBInterface
	FirstOrInit(dest interface{}, conds ...interface{}) *DBInterface
	Get(key string) (interface{}, bool)
	Group(name string) *DBInterface
	Having(query interface{}, args ...interface{}) *DBInterface
	InstanceGet(key string) (interface{}, bool)
	InstanceSet(key string, value interface{}) *DBInterface
	Joins(query string, args ...interface{}) *DBInterface
	Last(dest interface{}, conds ...interface{}) (tx *DBInterface)
	Limit(limit int) (tx *DBInterface)
	Migrator() gorm.Migrator
	Model(value interface{}) (tx *DBInterface)
	Not(query interface{}, args ...interface{}) (tx *DBInterface)
	Offset(offset int) (tx *DBInterface)
	Omit(columns ...string) (tx *DBInterface)
	Or(query interface{}, args ...interface{}) (tx *DBInterface)
	Order(value interface{}) (tx *DBInterface)
	Pluck(column string, dest interface{}) (tx *DBInterface)
	Preload(query string, args ...interface{}) (tx *DBInterface)
	Raw(sql string, values ...interface{}) (tx *DBInterface)
	Rollback() *DBInterface
	RollbackTo(name string) *DBInterface
	Row() *sql.Row
	Rows() (*sql.Rows, error)
	Save(value interface{}) (tx *DBInterface)
	SavePoint(name string) *DBInterface
	Scan(dest interface{}) (tx *DBInterface)
	ScanRows(rows *sql.Rows, dest interface{}) error
	Scopes(funcs ...func(*DBInterface) *DBInterface) (tx *DBInterface)
	Select(query interface{}, args ...interface{}) (tx *DBInterface)
	Session(config *gorm.Session) *DBInterface
	Set(key string, value interface{}) *DBInterface
	SetupJoinTable(model interface{}, field string, joinTable interface{}) error
	Table(name string, args ...interface{}) (tx *DBInterface)
	Take(dest interface{}, conds ...interface{}) (tx *DBInterface)
	ToSQL(queryFn func(tx *DBInterface) *DBInterface) string
	Transaction(fc func(tx *DBInterface) error, opts ...*sql.TxOptions) (err error)
	Unscoped() (tx *DBInterface)
	Update(column string, value interface{}) (tx *DBInterface)
	UpdateColumn(column string, value interface{}) (tx *DBInterface)
	UpdateColumns(values interface{}) (tx *DBInterface)
	Updates(values interface{}) (tx *DBInterface)
	Use(plugin gorm.Plugin) error
	Where(query interface{}, args ...interface{}) (tx *DBInterface)
	WithContext(ctx context.Context) *DBInterface
}

var PublicDB *gorm.DB
var PrivateDB *gorm.DB

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
	PublicDB, err = gorm.Open(mysql.Open(db_dns), &gorm.Config{})
	if err != nil {
		panic(err.Error())
	}
	tokens_db_dns := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local", tokens_db_user, tokens_db_pass, tokens_db_host, tokens_db_port, tokens_db_name)
	fmt.Printf("tokens_db_dns: %s", tokens_db_dns)
	PrivateDB, err = gorm.Open(mysql.Open(tokens_db_dns), &gorm.Config{})
	if err != nil {
		panic(err.Error())
	}
}
