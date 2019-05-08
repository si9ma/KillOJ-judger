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

type Product struct {
	ID        int         `gorm:"column:id;primary_key" json:"id"`
	CreatedAt null.Time   `gorm:"column:created_at" json:"created_at"`
	UpdatedAt null.Time   `gorm:"column:updated_at" json:"updated_at"`
	DeletedAt null.Time   `gorm:"column:deleted_at" json:"deleted_at"`
	Code      null.String `gorm:"column:code" json:"code"`
	Price     null.Int    `gorm:"column:price" json:"price"`
}

// TableName sets the insert table name for this struct type
func (p *Product) TableName() string {
	return "products"
}
