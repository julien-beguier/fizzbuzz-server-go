package controller

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-playground/validator"
	"github.com/julien-beguier/fizzbuzz-server-go/model"
	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

var DBgorm *gorm.DB

type QueryParam struct {
	Limit string `validate:"required,numeric"`
	Int1  string `validate:"required,numeric"`
	Int2  string `validate:"required,numeric"`
	Str1  string `validate:"required,alphanum,max=64"`
	Str2  string `validate:"required,alphanum,max=64"`
}

// Returns the fizzbuzz numbers according to the given mandatory parameters.
// Fizzbuzz algorithm is described as below:
//
// - limit: is the max number the algorithm will count to.
//
// - int1 & int2: are the numbers to look for multiples.
//
// - str1 & str2: are the strings to replace the numbers multiple of int1 & int2.
//
// Additional rule: if a number is a multiple of both int1 & int2, it will be
// replaced by the concatenation of str1 & str2.
//
// After every parameters are checked, considered valid and the numbers list
// processed, the parameters are saved in a MySQL database for further requesting.
//
// Errors:
//
// If one or more parameter is missing or incorrect, this method returns an error
// message with http error code 400 (Bad request).
func GetFizzbuzzNumbers(w http.ResponseWriter, r *http.Request) {
	queryParamMap := r.URL.Query()
	queryParam := QueryParam{
		Limit: queryParamMap.Get("limit"),
		Int1:  queryParamMap.Get("int1"),
		Int2:  queryParamMap.Get("int2"),
		Str1:  queryParamMap.Get("str1"),
		Str2:  queryParamMap.Get("str2")}

	stat, errParameters := CheckParams(queryParam)
	if errParameters != nil {
		// One or more parameter is not valid, abort with 400
		http.Error(w, errParameters.Error(), 400)
		return
	}

	// FIZZBUZZ ALGO
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
			// Not a multiple
			sb.WriteString(fmt.Sprint(i))
		}
	}

	// Check if the request (parameters) already exists in DB
	err := DBgorm.Where(stat).Take(stat).Error
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		log.Fatal(err)
	}

	// Increment the request hits counter before the insert or update
	stat.Hits++

	if err != nil && errors.Is(err, gorm.ErrRecordNotFound) {
		// Insert if the request isn't already saved
		if tx := DBgorm.Create(stat); tx.Error != nil {
			log.Fatal(tx.Error)
		}
	} else {
		// Or update the found record
		if tx := DBgorm.Save(stat); tx.Error != nil {
			log.Fatal(tx.Error)
		}
	}

	// Send the fizzbuzz numbers
	log.Info("Fizzbuzz numbers printed")
	writeToCaller(w, sb.String())
}

// Returns the most used parameters for the fizzbuzz numbers. It is possible
// to get multiples rows by calling this method.
//
// Errors:
//
// - Does not accept any parameter, if one or more parameter is provided, this
// method returns an error message with http error code 400 (Bad request).
//
// - If there isn't at least one statistic row saved in database this method
// returns an error message with http error code 404 (Not found).
func GetStatistics(w http.ResponseWriter, r *http.Request) {
	// Accept no parameter
	queryParamMap := r.URL.Query()
	if len(queryParamMap) != 0 {
		http.Error(w, "this endpoint does not accept parameter", 400)
		return
	}

	statistics := []model.Statistic{}
	// SELECT * FROM `statistic` WHERE hits = (SELECT MAX(hits) from `statistic`)
	if tx := DBgorm.Where("hits = (?)", DBgorm.Table("statistic").Select("MAX(hits)")).Find(&statistics); tx.Error != nil {
		log.Fatal(tx.Error)
	}

	// There is no row
	if len(statistics) == 0 {
		http.Error(w, "there isn't any saved request yet", 404)
	} else {
		// There is at least one row
		for i, stat := range statistics {
			writeToCaller(w, fmt.Sprintf("Request n°%d : %s", i+1, stat.ToString()))
		}
	}
}

// Writes to the caller with http code 200 (OK)
func writeToCaller(w http.ResponseWriter, s string) {
	_, e := fmt.Fprintln(w, s)
	if e != nil {
		log.WithError(e).Fatal("unexpected error")
	}
}

// Builds a string containing all error messages separated by a new line '\n'
func buildErrorMessage(errorString string, s string) string {
	if len(errorString) > 0 {
		return fmt.Sprintf("%s\n%s", errorString, s)
	} else {
		return s
	}
}

// Checks to see if an int is valid for the fizzbuzz algorithm.
//
// Uses: strconv.Atoi(string)
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

// Checks & validates the given parameters for the fizzbuzz algorithm.
//
// In case of error(s), builds an error with all detected errors and returns it.
func CheckParams(qp QueryParam) (*model.Statistic, error) {
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

	s := model.Statistic{}

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
