package model

type UserDetails struct {
	UserId     int64
	Username   string
	Password   string
	Autorities []string
}

func (user *UserDetails) IsMatch(username, password string) bool {
	return user.Username == username && user.Password == password
}
