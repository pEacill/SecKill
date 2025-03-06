package mysql

import (
	"fmt"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var db *gorm.DB
var err error

func InitMysql(hostMysql, portMysql, userMysql, pwdMysql, dbMysql string) {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8&parseTime=True&loc=Local",
		userMysql, pwdMysql, hostMysql, portMysql, dbMysql)

	db, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		fmt.Println(err)
		return
	}

	sqlDB, err := db.DB()
	if err != nil {
		fmt.Println(err)
		return
	}
	sqlDB.SetMaxOpenConns(300)
	sqlDB.SetMaxIdleConns(10)
}

func DB() *gorm.DB {
	return db
}
