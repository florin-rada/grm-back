package jobs

import (
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

// Define a test suite
type JobsModelSuite struct {
	suite.Suite
	dbMock sqlmock.Sqlmock
	model  JobsModel
}

// Setup the test suite
func (suite *JobsModelSuite) SetupTest() {
	// Initialize the SQL mock
	db, mock, _ := sqlmock.New()
	suite.dbMock = mock
	db_user := os.Getenv("DB_USER")
	//db_pass := os.Getenv("DB_PASS")
	//db_host := os.Getenv("DB_HOST")
	//db_port := os.Getenv("DB_PORT")
	//db_name := os.Getenv("DB_NAME")
	if db_user == "" {
		panic("WTF We have no user")
	}
	dns := "root:root@tcp(127.0.0.1:3306)/grm?charset=utf8mb4&parseTime=True&loc=Local"
	fmt.Println(dns)
	dialector := mysql.New(mysql.Config{
		DSN:                       dns,
		DriverName:                "mysql",
		Conn:                      db,
		SkipInitializeWithVersion: true,
	})
	gormDB, _ := gorm.Open(dialector)
	// Initialize the JobsModel with the mocked DB and client
	suite.model = JobsModel{
		db:     gormDB,
		client: nil, // Replace with your GitLab client initialization
	}
}

// Define the tests
func (suite *JobsModelSuite) TestGetRunnerJobs() {
	// Prepare some sample data
	idRunner := uint(1)
	idUser := "user123"
	page := 1
	perPage := 10

	// Set up expectations for the mocked DB
	suite.dbMock.ExpectQuery("SELECT(.*)").
		WithArgs(idRunner, idUser).
		WillReturnRows(sqlmock.NewRows([]string{
			"internal_user_id", "id", "name", "created_at", "started_at", "finished_at",
			"pipeline_id", "project_id", "runner_id", "branch", "duration", "queued_duration",
			"url", "stage",
		}).
			AddRow("user123", 1, "Job 1", time.Now(), time.Now(), time.Now(), 1, 1, 1, "branch", 10.0, 5.0, "url1", "stage1").
			AddRow("user123", 2, "Job 2", time.Now(), time.Now(), time.Now(), 2, 2, 1, "branch", 15.0, 6.0, "url2", "stage2"))

	// Call the function being tested
	jobs, err := suite.model.GetRunnerJobs(idRunner, idUser, page, perPage)

	// Assert the results
	assert.NoError(suite.T(), err)
	assert.Len(suite.T(), jobs, 2)
	assert.Equal(suite.T(), "Job 1", jobs[0].Name)
	assert.Equal(suite.T(), "Job 2", jobs[1].Name)

	// Ensure all expectations were met
	suite.dbMock.ExpectationsWereMet()
}

// Run the test suite
func TestJobsModelSuite(t *testing.T) {
	suite.Run(t, new(JobsModelSuite))
}
