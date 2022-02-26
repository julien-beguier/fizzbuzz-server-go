package main

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"time"

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

type DatabaseVars struct {
	user     string
	password string
	database string
	host     string
	port     string
}

// Retrieve the database connection informations from the environnement and returns
// a struct containing the values as strings
func getDatabaseVariablesFromEnv() DatabaseVars {
	userString, userBool := os.LookupEnv("DATABASE_USER")
	passwordString, passwordBool := os.LookupEnv("DATABASE_PASS")
	databaseString, databaseBool := os.LookupEnv("DATABASE_NAME")
	hostString, hostBool := os.LookupEnv("DATABASE_HOST")
	portString, portBool := os.LookupEnv("DATABASE_PORT")

	if !userBool || !passwordBool || !databaseBool || !hostBool || !portBool {
		log.Fatal("failed to retrieve the database informations from environnement")
	}

	dbVars := DatabaseVars{
		user:     userString,
		password: passwordString,
		database: databaseString,
		host:     hostString,
		port:     portString,
	}

	return dbVars
}

// Try to connect to the database and sets the Gorm object if it succeed.
//
// If there is an error, the program will abort.
func dbConnect(dbVars DatabaseVars) *gorm.DB {
	dsn := dbVars.user + ":" + dbVars.password + "@tcp(" + dbVars.host + ":" + dbVars.port + ")/" + dbVars.database + "?charset=utf8mb4&parseTime=True&loc=Local"

	// Interval at which a new try is done, 5 seconds
	ticker := time.NewTicker(time.Second * 10)
	// Timeout of 5 minutes for mysql initialization
	timeout := time.NewTicker(time.Minute * 5)
	for {
		select {
		// If timeout is reached, abort
		case <-timeout.C:
			log.Fatal(errors.New("failed to connect to database"))
		case <-ticker.C:
			if db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{}); err == nil {
				return db
			}
		}
	}
}

func main() {

	// Retrieve the database information from environnement
	dbVars := getDatabaseVariablesFromEnv()
	// Database initialization first
	DBgorm := dbConnect(dbVars)

	// Get the sql.DB to close the connection later
	sqlDB, err := DBgorm.DB()
	if err != nil {
		log.Fatal("failed to retrieve the sql.DB from gorm", err)
	}
	defer sqlDB.Close()

	// Service needs to perform requests, but controller will pass it
	// to the service
	controller := controller.NewController(DBgorm)

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
