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

type Group struct {
}

// TableName sets the insert table name for this struct type
func (g *Group) TableName() string {
	return "group"
}
