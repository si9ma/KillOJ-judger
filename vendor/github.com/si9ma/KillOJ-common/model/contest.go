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

type Contest struct {
	ID          int       `gorm:"column:id;primary_key" json:"id"`
	ContestName string    `gorm:"column:contest_name" json:"contest_name"`
	CreateTime  time.Time `gorm:"column:create_time" json:"create_time"`
	Owner       int       `gorm:"column:owner" json:"owner"`
	ManagerList string    `gorm:"column:manager_list" json:"manager_list"`
	AllowGroups string    `gorm:"column:allow_groups" json:"allow_groups"`
	AllowUsers  string    `gorm:"column:allow_users" json:"allow_users"`
	StartTime   time.Time `gorm:"column:start_time" json:"start_time"`
	EndTime     time.Time `gorm:"column:end_time" json:"end_time"`
}

// TableName sets the insert table name for this struct type
func (c *Contest) TableName() string {
	return "contest"
}
