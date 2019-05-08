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

type UserInContest struct {
	ID        int       `gorm:"column:id;primary_key" json:"id"`
	ContestID int       `gorm:"column:contest_id" json:"contest_id"`
	UserID    int       `gorm:"column:user_id" json:"user_id"`
	JoinTime  time.Time `gorm:"column:join_time" json:"join_time"`
}

// TableName sets the insert table name for this struct type
func (u *UserInContest) TableName() string {
	return "user_in_contest"
}
