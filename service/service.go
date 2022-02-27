package service

import (
	"errors"
	"fmt"
	"strings"

	"github.com/julien-beguier/fizzbuzz-server-go/model"
	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type Service struct {
	DBgorm *gorm.DB
}

func NewService(DBgorm *gorm.DB) *Service {
	return &Service{
		DBgorm: DBgorm,
	}
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
func (s *Service) FizzbuzzList(stat *model.Statistic) string {

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
	return sb.String()
}

// Save the parameters in DB.
//
// Perform either an insert if those were not inserted before or an update.
func (s *Service) InsertOrUpdateStatistic(stat *model.Statistic) {
	// Check if the request (parameters) already exists in DB
	err := s.DBgorm.Where(stat).Take(stat).Error
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		log.Fatal(err)
	}

	// Increment the request hits counter before the insert or update
	stat.Hits++

	if err != nil && errors.Is(err, gorm.ErrRecordNotFound) {
		// Insert if the request isn't already saved
		if tx := s.DBgorm.Create(stat); tx.Error != nil {
			log.Fatal(tx.Error)
		}
	} else {
		// Or update the found record
		if tx := s.DBgorm.Save(stat); tx.Error != nil {
			log.Fatal(tx.Error)
		}
	}
}

// Returns the most used parameters for the fizzbuzz numbers. It is possible
// to get multiples rows by calling this method.
func (s *Service) RetrieveStatistics() []model.Statistic {
	statistics := []model.Statistic{}

	// SELECT * FROM `statistic` WHERE hits = (SELECT MAX(hits) from `statistic`)
	if tx := s.DBgorm.Where("hits = (?)", s.DBgorm.Table("statistic").Select("MAX(hits)")).Find(&statistics); tx.Error != nil {
		log.Fatal(tx.Error)
	}

	return statistics
}
