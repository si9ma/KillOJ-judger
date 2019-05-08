package model

import (
	"database/sql"
	"time"

	"github.com/guregu/null"
)

var (
	_ = time.Second
	_ = sql.LevelDefault
	_ = null.Bool{}
)

type Statistic struct {
	ID           int       `gorm:"column:id;primary_key" json:"id"`
	Date         time.Time `gorm:"column:date" json:"date"`
	UserCount    int       `gorm:"column:user_count" json:"user_count"`
	TagCount     int       `gorm:"column:tag_count" json:"tag_count"`
	CataCount    int       `gorm:"column:cata_count" json:"cata_count"`
	SubmitCount  int       `gorm:"column:submit_count" json:"submit_count"`
	ProblemCount int       `gorm:"column:problem_count" json:"problem_count"`
	GroupCount   int       `gorm:"column:group_count" json:"group_count"`
	ContestCount int       `gorm:"column:contest_count" json:"contest_count"`
}

// TableName sets the insert table name for this struct type
func (s *Statistic) TableName() string {
	return "statistics"
}
