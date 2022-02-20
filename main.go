package main

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-playground/validator"
	log "github.com/sirupsen/logrus"
)

const (
	PORT = 8080
)

type QueryParam struct {
	Limit string `validate:"required,numeric"`
	Int1  string `validate:"required,numeric"`
	Int2  string `validate:"required,numeric"`
	Str1  string `validate:"required,alphanum,max=64"`
	Str2  string `validate:"required,alphanum,max=64"`
}

// Struct saved in DB
type Statistic struct {
	gorm.Model
	Limit uint
	Int1  uint
	Int2  uint
	Str1  string
	Str2  string
	Hits  uint
}

var DBgorm *gorm.DB

func (Statistic) TableName() string {
	return "statistic"
}

func (s *Statistic) ToString() string {
	return fmt.Sprintf(
		"limit=%d, int1=%d, int2=%d, str1=%s, str2=%s, hits=%d, created_at=%s, updated_at=%s\n",
		s.Limit,
		s.Int1,
		s.Int2,
		s.Str1,
		s.Str2,
		s.Hits,
		s.CreatedAt.String(),
		s.UpdatedAt.String())
}

func dbConnect() {
	dbUser := "fizzbuzz-user"
	dbPass := "7bMP+_qjyyAVy+=mY+DU"
	dbName := "fizzbuzz"
	dsn := dbUser + ":" + dbPass + "@tcp(127.0.0.1:3306)/" + dbName + "?charset=utf8mb4&parseTime=True&loc=Local"
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})

	if err != nil {
		log.Fatal("failed to connect to database", err)
	}
	DBgorm = db
}

func writeToCaller(w http.ResponseWriter, s string) {
	_, e := fmt.Fprintln(w, s)
	if e != nil {
		log.WithError(e).Fatal("unexpected error")
	}
}

func buildErrorMessage(errorString string, s string) string {
	if len(errorString) > 0 {
		return fmt.Sprintf("%s\n%s", errorString, s)
	} else {
		return s
	}
}

func checkIntFromParam(s string) (int, error) {
	v, err := strconv.Atoi(s)

	//strconv.ErrSyntax is handled by validator
	if err != nil && errors.Is(err, strconv.ErrSyntax) {
		return 0, nil
	} else if err != nil && errors.Is(err, strconv.ErrRange) {
		err = fmt.Errorf("int type parameter is out of range (received:%s)", s)
		return -1, err
	} else if v < 1 {
		err = fmt.Errorf("int type parameter cannot be less than 1 (received:%s)", s)
		return -1, err
	}
	return v, nil
}

func checkParams(qp QueryParam) (*Statistic, error) {
	errorString := ""
	validate := validator.New()

	err := validate.Struct(qp)
	if err != nil {
		if _, ko := err.(*validator.InvalidValidationError); ko {
			log.WithError(err).Fatal("unexpected error")
		}

		for _, err := range err.(validator.ValidationErrors) {
			if err.ActualTag() == "required" {
				errorString = buildErrorMessage(errorString, fmt.Sprintf("parameter %s is required", strings.ToLower(err.StructField())))
			} else if err.ActualTag() == "numeric" {
				errorString = buildErrorMessage(errorString, fmt.Sprintf("parameter %s is not a numeric value (received:%s)", strings.ToLower(err.StructField()), err.Value()))
			} else if err.ActualTag() == "alphanum" {
				errorString = buildErrorMessage(errorString, fmt.Sprintf("parameter %s is not an alphanumeric value (received:%s)", strings.ToLower(err.StructField()), err.Value()))
			} else if err.ActualTag() == "max" {
				errorString = buildErrorMessage(errorString, fmt.Sprintf("parameter %s cannot be over 64 characters (received:%s)", strings.ToLower(err.StructField()), err.Value()))
			}
		}
	}

	s := Statistic{}

	v, err := checkIntFromParam(qp.Limit)
	if err != nil {
		errorString = buildErrorMessage(errorString, err.Error())
	}
	s.Limit = uint(v)

	v, err = checkIntFromParam(qp.Int1)
	if err != nil {
		errorString = buildErrorMessage(errorString, err.Error())
	}
	s.Int1 = uint(v)

	v, err = checkIntFromParam(qp.Int2)
	if err != nil {
		errorString = buildErrorMessage(errorString, err.Error())
	}
	s.Int2 = uint(v)

	// Validator already checked the strings parameters
	s.Str1 = qp.Str1
	s.Str2 = qp.Str2

	if len(errorString) > 0 {
		return nil, fmt.Errorf(errorString)
	}
	return &s, nil
}

func init() {
	dbConnect()
}

func main() {

	sqlDB, err := DBgorm.DB()
	if err != nil {
		log.Fatal(err)
	}
	defer sqlDB.Close()

	// On the first launch, will initialize the db & create tables, fields, keys, indexes ...
	DBgorm.AutoMigrate(&Statistic{})

	r := chi.NewRouter()
	r.Use(middleware.Logger)

	// Route to get the fizzbuzz numbers
	r.Get("/list", getFizzbuzzNumbers)

	// Route to get the statistics
	r.Get("/statistics", getStatistics)

	log.WithField("port", PORT).Info("Serving http server...")
	if err := http.ListenAndServe(fmt.Sprintf(":%d", PORT), r); err != nil {
		log.WithError(err).Fatal("unexpected error")
	}
}

func getFizzbuzzNumbers(w http.ResponseWriter, r *http.Request) {
	queryParamMap := r.URL.Query()
	queryParam := QueryParam{
		Limit: queryParamMap.Get("limit"),
		Int1:  queryParamMap.Get("int1"),
		Int2:  queryParamMap.Get("int2"),
		Str1:  queryParamMap.Get("str1"),
		Str2:  queryParamMap.Get("str2")}

	stat, errParameters := checkParams(queryParam)
	if errParameters != nil {
		// One or more parameter is not valid, abort with 400
		http.Error(w, errParameters.Error(), 400)
		return
	}

	// FIZZ BUZZ ALGO
	strBoth := stat.Str1 + stat.Str2
	sb := strings.Builder{}

	for i := uint(1); i <= stat.Limit; i++ {
		if i != 1 {
			sb.WriteString(", ")
		}

		if i%stat.Int1 == 0 && i%stat.Int2 == 0 {
			// Multiples of both int1 & int2
			sb.WriteString(strBoth)
		} else if i%stat.Int2 == 0 {
			// Multiples of both int2
			sb.WriteString(stat.Str2)
		} else if i%stat.Int1 == 0 {
			// Multiples of both int1
			sb.WriteString(stat.Str1)
		} else {
			sb.WriteString(fmt.Sprint(i))
		}
	}

	// Check if the request already exists in DB
	err := DBgorm.Where(stat).Take(stat).Error
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		log.Fatal(err)
	}
	// Increment the request hits counter
	stat.Hits++
	if err != nil && errors.Is(err, gorm.ErrRecordNotFound) {
		// Insert if the request isn't already saved
		DBgorm.Create(stat) // TODO error check
	} else {
		// Or Update the found record
		DBgorm.Save(stat) // TODO error check
	}

	// Send the fizzbuzz numbers
	log.Info("Fizzbuzz numbers printed")
	writeToCaller(w, sb.String())
}

func getStatistics(w http.ResponseWriter, r *http.Request) {

	// Accept no parameter
	queryParamMap := r.URL.Query()
	if len(queryParamMap) != 0 {
		acceptNoParameter := "this endpoint does not accept parameter"
		log.Error(acceptNoParameter)
		http.Error(w, acceptNoParameter, 400)
		return
	}

	statistics := []Statistic{}
	// SELECT * FROM `statistic` WHERE hits= (SELECT MAX(hits) from `statistic`)
	DBgorm.Where("hits = (?)", DBgorm.Table("statistic").Select("MAX(hits)")).Find(&statistics) // TODO error check

	// There is no row
	if len(statistics) == 0 {
		noStatToPrint := "there isn't any saved request yet."
		http.Error(w, noStatToPrint, 404)
	} else {
		// There is at least one row
		i := 1
		for _, stat := range statistics {
			stringStat := stat.ToString()
			writeToCaller(w, fmt.Sprintf("Request n°%d : %s", i, stringStat))
			i++
		}
	}
}
