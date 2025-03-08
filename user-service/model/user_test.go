package model

import (
	"testing"

	"github.com/pEacill/SecKill/pkg/mysql"
)

var (
	hostMysql = "localhost"
	portMysql = "3306"
	userMysql = "root"
	pwdMysql  = "root"
	dbMysql   = "user"
)

func TestCheckUser(t *testing.T) {
	mysql.InitMysql(hostMysql, portMysql, userMysql, pwdMysql, dbMysql)
	if mysql.DB() == nil {
		t.Fatal("Database connection failed")
	}

	userModel := NewUserModel()

	user, err := userModel.CheckUser("user", "password")
	if err != nil {
		t.Errorf("Error: %v", err)
	}
	if user.UserName != "user" {
		t.Errorf("not match!")
	}

	users, err := userModel.GetUserList()
	if err != nil {
		t.Errorf("GetUserList error: %v", err)
	}
	if len(users) == 0 {
		t.Log("No users found in database")
	} else {
		t.Logf("Found %d users in database", len(users))
	}

	newUser := &User{
		UserName: "testuser",
		Password: "testpassword",
		Age:      25,
	}

	err = userModel.CreateUser(newUser)
	if err != nil {
		t.Errorf("CreateUser error: %v", err)
	}

	createdUser, err := userModel.CheckUser("testuser", "testpassword")
	if err != nil {
		t.Errorf("Error verifying created user: %v", err)
	}
	if createdUser.UserName != "testuser" || createdUser.Password != "testpassword" || createdUser.Age != 25 {
		t.Errorf("Created user data does not match input")
	}
}
