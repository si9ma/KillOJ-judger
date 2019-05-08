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

type Owner struct {
	ID            int    `gorm:"column:id;primary_key" json:"id"`
	UserID        int    `gorm:"column:user_id" json:"user_id"`
	ProblemID     int    `gorm:"column:problem_id" json:"problem_id"`
	ManagerIDList string `gorm:"column:manager_id_list" json:"manager_id_list"`
}

// TableName sets the insert table name for this struct type
func (o *Owner) TableName() string {
	return "owner"
}
