package model

import (
	"time"
)

type Comment struct {
	ID        int       `gorm:"column:id;primary_key" json:"id"`
	CreatedAt time.Time `gorm:"column:created_at" json:"created_at"`
	UpdatedAt time.Time `gorm:"column:updated_at" json:"updated_at"`
	UserID    int       `gorm:"column:user_id" json:"user_id"`
	ProblemID int       `gorm:"column:problem_id" json:"problem_id"`
	ParentID  int       `gorm:"column:parent_id" json:"parent_id"`
	Content   string    `gorm:"column:content" json:"content"`
	User      User      `json:"user"`
}

// TableName sets the insert table name for this struct type
func (d *Comment) TableName() string {
	return "comment"
}
