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

type Discuss struct {
	DiscussID       int    `gorm:"column:discuss_id;primary_key" json:"discuss_id"`
	ProblemID       int    `gorm:"column:problem_id" json:"problem_id"`
	FromUserID      int    `gorm:"column:from_user_id" json:"from_user_id"`
	ToUserID        int    `gorm:"column:to_user_id" json:"to_user_id"`
	Content         string `gorm:"column:content" json:"content"`
	LeaderDiscussID int    `gorm:"column:leader_discuss_id" json:"leader_discuss_id"`
	Ok              int    `gorm:"column:ok" json:"ok"`
	NoOk            int    `gorm:"column:no_ok" json:"no_ok"`
	OkList          string `gorm:"column:ok_list" json:"ok_list"`
	NoOkList        string `gorm:"column:no_ok_list" json:"no_ok_list"`
}

// TableName sets the insert table name for this struct type
func (d *Discuss) TableName() string {
	return "discuss"
}
