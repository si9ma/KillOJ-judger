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

type Catalog struct {
	ID         int       `gorm:"column:id;primary_key" json:"id"`
	CataName   string    `gorm:"column:cata_name" json:"cata_name"`
	Desc       string    `gorm:"column:desc" json:"desc"`
	CreateTime time.Time `gorm:"column:create_time" json:"create_time"`
}

// TableName sets the insert table name for this struct type
func (c *Catalog) TableName() string {
	return "catalog"
}
