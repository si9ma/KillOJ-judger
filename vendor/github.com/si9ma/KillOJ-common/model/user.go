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

type User struct {
	ID           int       `gorm:"column:id;primary_key" json:"id"`
	CreateAt     time.Time `gorm:"column:create_at" json:"create_at"`
	UpdateAt     time.Time `gorm:"column:update_at" json:"update_at"`
	NickName     string    `gorm:"column:nick_name" json:"nick_name"`
	Signature    string    `gorm:"column:signature" json:"signature"`
	StudentNum   string    `gorm:"column:student_num" json:"student_num"`
	Organization string    `gorm:"column:organization" json:"organization"`
	Email        string    `gorm:"column:email" json:"email"`
	Site         string    `gorm:"column:site" json:"site"`
	GithubID     string    `gorm:"column:github_id" json:"github_id"`
}

// TableName sets the insert table name for this struct type
func (u *User) TableName() string {
	return "user"
}
