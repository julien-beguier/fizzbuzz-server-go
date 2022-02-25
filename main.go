package main

import (
	"fmt"
	"net/http"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/julien-beguier/fizzbuzz-server-go/controller"
	"github.com/julien-beguier/fizzbuzz-server-go/model"
	log "github.com/sirupsen/logrus"
)

const (
	PORT = 8080
)

var DBgorm *gorm.DB

// Try to connect to the database and sets the Gorm object if it succeed.
//
// If there is an error, the program will abort.
func dbConnect() {
	dbUser := "fizzbuzz-user"
	dbPass := "7bMP+_qjyyAVy+=mY+DU"
	dbName := "fizzbuzz"
	dsn := dbUser + ":" + dbPass + "@tcp(:3306)/" + dbName + "?charset=utf8mb4&parseTime=True&loc=Local"

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("failed to connect to database", err)
	}

	DBgorm = db
}

func init() {
	// Database initialisation first
	dbConnect()
}

func main() {

	sqlDB, err := DBgorm.DB()
	if err != nil {
		log.Fatal("failed to retrieve the sql.DB from gorm", err)
	}
	defer sqlDB.Close()
	// Controller needs to perform requests
	controller.DBgorm = DBgorm

	// On the first launch, will initialize the db & create tables, fields, keys, indexes
	if err = DBgorm.AutoMigrate(&model.Statistic{}); err != nil {
		log.Fatal("failed to auto migrate the database using the given model", err)
	}

	r := chi.NewRouter()
	r.Use(middleware.Logger)

	// Route to get the fizzbuzz numbers
	r.Get("/list", controller.GetFizzbuzzNumbers)

	// Route to get the statistics
	r.Get("/statistics", controller.GetStatistics)

	log.WithField("port", PORT).Info("Serving http server...")
	if err := http.ListenAndServe(fmt.Sprintf(":%d", PORT), r); err != nil {
		log.WithError(err).Fatal("unexpected error")
	}
}
