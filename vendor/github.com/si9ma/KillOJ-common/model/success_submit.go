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

type SuccessSubmit struct {
	ID          int       `gorm:"column:id;primary_key" json:"id"`
	CreatedAt   time.Time `gorm:"column:created_at" json:"created_at"`
	UpdateAt    time.Time `gorm:"column:update_at" json:"update_at"`
	UserID      int       `gorm:"column:user_id" json:"user_id"`
	ProblemID   int       `gorm:"column:problem_id" json:"problem_id"`
	SourceCode  string    `gorm:"column:source_code" json:"source_code"`
	Language    int       `gorm:"column:language" json:"language"`
	Result      int       `gorm:"column:result" json:"result"`
	RunTime     int       `gorm:"column:run_time" json:"run_time"`
	MemoryUsage int       `gorm:"column:memory_usage" json:"memory_usage"`
}

// TableName sets the insert table name for this struct type
func (s *SuccessSubmit) TableName() string {
	return "success_submit"
}
