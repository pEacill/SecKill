package model

import (
	"log"

	"github.com/pEacill/SecKill/pkg/mysql"
	"gorm.io/gorm"
)

type User struct {
	UserId   int64  `json:"user_id" gorm:"column:user_id;primaryKey"`
	UserName string `json:"user_name" gorm:"column:user_name"`
	Password string `json:"password" gorm:"column:password"`
	Age      int    `json:"age"`
}

type UserModel struct {
	DB *gorm.DB
}

func NewUserModel() *UserModel {
	return &UserModel{
		DB: mysql.DB(),
	}
}

func (u *UserModel) getTableName() string {
	return "user"
}

func (u *UserModel) GetUserList() ([]User, error) {
	var users []User
	err := u.DB.Table(u.getTableName()).Find(&users).Error
	if err != nil {
		log.Printf("Error Get Users List: %v", err)
	}
	return users, nil
}

func (u *UserModel) CheckUser(username, password string) (*User, error) {
	var user User
	err := u.DB.Table(u.getTableName()).Where("user_name = ? AND password = ?", username, password).
		First(&user).
		Error
	if err != nil {
		log.Printf("Error checking user: %v", err)
		return nil, err
	}

	return &user, nil
}

func (u *UserModel) CreateUser(user *User) error {
	err := u.DB.Table(u.getTableName()).Create(user).Error
	if err != nil {
		log.Printf("Error creating user: %v", err)
		return err
	}

	return nil
}
