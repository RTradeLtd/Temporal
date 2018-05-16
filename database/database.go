package database

import (
	"log"

	"github.com/RTradeLtd/Temporal/models"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
)

/*
	roll.Token = "POST_SERVER_ITEM_ACCESS_TOKEN"
	//roll.Environment = "production" // defaults to "development"

	r := gin.Default()
	r.Use(rollbar.Recovery(true))

	r.Run(":8080")
	func l(err error) {
	token := os.Getenv("ROLLBAR_TOKEN")
	rollbar.SetToken(token)
	rollbar.SetServerRoot("github.com/RTradeLtd/Temporal") // path of project (required for GitHub integration and non-project stacktrace collapsing)

	rollbar.Error(err)

	rollbar.Wait()
}
*/

var uploadObj *models.Upload
var userObj *models.User

type DatabaseManager struct {
	DB     *gorm.DB
	Upload *models.UploadManager
}

func Initialize() *DatabaseManager {
	dbm := DatabaseManager{}
	db := OpenDBConnection()
	dbm.DB = db
	dbm.RunMigrations()
	return &dbm
}

func (dbm *DatabaseManager) RunMigrations() {
	dbm.DB.AutoMigrate(uploadObj)
	dbm.DB.AutoMigrate(userObj)
}

// OpenDBConnection is used to create a database connection
func OpenDBConnection() *gorm.DB {
	db, err := gorm.Open("sqlite3", "./ipfs_database.db")
	if err != nil {
		log.Fatal(err)
	}
	return db
}

// CloseDBConnection is used to close a db
func CloseDBConnection(db *gorm.DB) {
	db.Close()
}
