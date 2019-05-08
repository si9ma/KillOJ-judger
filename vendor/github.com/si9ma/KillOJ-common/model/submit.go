package model

import (
	"time"
)

type Submit struct {
	ID          int       `gorm:"column:id;primary_key" json:"id"`
	CreatedAt   time.Time `gorm:"column:created_at" json:"created_at"`
	UpdateAt    time.Time `gorm:"column:update_at" json:"update_at"`
	ProblemID   int       `gorm:"column:problem_id" json:"problem_id"`
	UserID      int       `gorm:"column:user_id" json:"user_id"`
	SourceCode  string    `gorm:"column:source_code" json:"source_code"`
	Language    int       `gorm:"column:language" json:"language"`
	Result      int       `gorm:"column:result" json:"result"`
	RunTime     int       `gorm:"column:run_time" json:"run_time"`
	MemoryUsage int       `gorm:"column:memory_usage" json:"memory_usage"`
	IsComplete  int       `gorm:"column:is_complete" json:"is_complete"`
	Problem     Problem
}

// TableName sets the insert table name for this struct type
func (s *Submit) TableName() string {
	return "submit"
}
