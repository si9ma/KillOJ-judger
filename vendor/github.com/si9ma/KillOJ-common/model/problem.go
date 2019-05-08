package model

import (
	"time"
)

type Problem struct {
	ID              int       `gorm:"column:id;primary_key" json:"id"`
	CreateAt        time.Time `gorm:"column:create_at" json:"create_at"`
	UpdateAt        time.Time `gorm:"column:update_at" json:"update_at"`
	Desc            string    `gorm:"column:desc" json:"desc"`
	TimeLimit       int       `gorm:"column:time_limit" json:"time_limit"`
	MemoryLimit     int       `gorm:"column:memory_limit" json:"memory_limit"`
	DifficultyCata  int       `gorm:"column:difficulty_cata" json:"difficulty_cata"`
	Difficulty      int       `gorm:"column:difficulty" json:"difficulty"`
	Ok              int       `gorm:"column:ok" json:"ok"`
	NoOk            int       `gorm:"column:no_ok" json:"no_ok"`
	OkList          string    `gorm:"column:ok_list" json:"ok_list"`
	NoOkList        string    `gorm:"column:no_ok_list" json:"no_ok_list"`
	Private         int       `gorm:"column:private" json:"private"`
	BelongTo        int       `gorm:"column:belong_to" json:"belong_to"`
	BelongType      int       `gorm:"column:belong_type" json:"belong_type"`
	CataID          int       `gorm:"column:cata_id" json:"cata_id"`
	ProblemTestCase []ProblemTestCase
}

// TableName sets the insert table name for this struct type
func (p *Problem) TableName() string {
	return "problem"
}
