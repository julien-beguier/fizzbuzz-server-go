package model

import (
	"fmt"

	"gorm.io/gorm"
)

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
