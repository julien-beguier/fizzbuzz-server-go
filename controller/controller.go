package controller

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-playground/validator"
	"github.com/julien-beguier/fizzbuzz-server-go/model"
	"github.com/julien-beguier/fizzbuzz-server-go/service"
	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type QueryParam struct {
	Limit string `validate:"required,numeric"`
	Int1  string `validate:"required,numeric"`
	Int2  string `validate:"required,numeric"`
	Str1  string `validate:"required,alphanum,max=64"`
	Str2  string `validate:"required,alphanum,max=64"`
}

type Controller struct {
	service *service.Service
}

func NewController(DBgorm *gorm.DB) *Controller {
	return &Controller{
		service: service.NewService(DBgorm),
	}
}

// Check parameters, call the service to get the fizzbuzz numbers list.
// Then the service is called to save the parameters in a MySQL database for
// further requesting.
//
// Errors:
//
// If one or more parameter is missing or incorrect, this method returns an error
// message with http error code 400 (Bad request).
func (c *Controller) GetFizzbuzzNumbers(w http.ResponseWriter, r *http.Request) {
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

	// Fizzbuzz Algo
	fizzbuzzNumbers := c.service.FizzbuzzList(stat)

	// Save the parameters in DB
	c.service.UpdateDB(stat)

	// Send the fizzbuzz numbers
	log.Info("Fizzbuzz numbers printed")
	writeToCaller(w, fizzbuzzNumbers)
}

// Call the service to get the most used parameters for the fizzbuzz numbers. It
// is possible to get multiples rows by calling this method.
//
// Errors:
//
// - Does not accept any parameter, if one or more parameter is provided, this
// method returns an error message with http error code 400 (Bad request).
//
// - If there isn't at least one statistic row saved in database this method
// returns an error message with http error code 404 (Not found).
func (c *Controller) GetStatistics(w http.ResponseWriter, r *http.Request) {
	// Accept no parameter
	queryParamMap := r.URL.Query()
	if len(queryParamMap) != 0 {
		http.Error(w, "this endpoint does not accept parameter", 400)
		return
	}

	// Query DB about the most used paramters
	statistics := c.service.RetriveStatistics()

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
		s = fmt.Sprintf("%s\n%s", errorString, s)
	}
	return s
}

// Checks to see if an int is valid for the fizzbuzz algorithm.
//
// Uses: strconv.Atoi(string)
func checkIntFromParam(s string) (int, error) {
	v, err := strconv.Atoi(s)

	if err != nil {
		// strconv.ErrSyntax is handled by validator
		if errors.Is(err, strconv.ErrSyntax) {
			return 0, nil
		} else if errors.Is(err, strconv.ErrRange) {
			err = fmt.Errorf("int type parameter is out of range (received:%s)", s)
			return -1, err
		}
	}

	if v < 1 {
		err = fmt.Errorf("int type parameter cannot be less than 1 (received:%s)", s)
		return -1, err
	}
	return v, nil
}

// Checks & validates the given parameters for the fizzbuzz algorithm.
//
// In case of error(s), builds an error with all detected errors and returns it.
func checkParams(qp QueryParam) (*model.Statistic, error) {
	errorString := ""
	validate := validator.New()

	err := validate.Struct(qp)
	if err != nil {
		if _, ko := err.(*validator.InvalidValidationError); ko {
			log.WithError(err).Fatal("unexpected error")
		}

		for _, err := range err.(validator.ValidationErrors) {
			switch err.ActualTag() {
			case "required":
				errorString = buildErrorMessage(errorString, fmt.Sprintf("parameter %s is required", strings.ToLower(err.StructField())))
			case "numeric":
				errorString = buildErrorMessage(errorString, fmt.Sprintf("parameter %s is not a numeric value (received:%s)", strings.ToLower(err.StructField()), err.Value()))
			case "alphanum":
				errorString = buildErrorMessage(errorString, fmt.Sprintf("parameter %s is not an alphanumeric value (received:%s)", strings.ToLower(err.StructField()), err.Value()))
			case "max":
				errorString = buildErrorMessage(errorString, fmt.Sprintf("parameter %s cannot be over 64 characters (received:%s)", strings.ToLower(err.StructField()), err.Value()))
			default:
				// All possible cases handled by the validators
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
