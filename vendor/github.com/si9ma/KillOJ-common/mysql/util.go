package mysql

import "github.com/jinzhu/gorm"

func GetTestDB() (db *gorm.DB, err error) {
	db, err = gorm.Open("mysql", "root:mysqlpass@tcp(127.0.0.1:3306)/killoj?&parseTime=True")
	return
}
