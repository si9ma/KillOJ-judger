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

type UserInGroup struct {
	ID       int       `gorm:"column:id;primary_key" json:"id"`
	GroupID  int       `gorm:"column:group_id" json:"group_id"`
	UserID   int       `gorm:"column:user_id" json:"user_id"`
	JoinTime time.Time `gorm:"column:join_time" json:"join_time"`
}

// TableName sets the insert table name for this struct type
func (u *UserInGroup) TableName() string {
	return "user_in_group"
}
