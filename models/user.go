package models

type User struct {
	ID       int    `json:"id" gorm:"primary_key;auto_increment"`
	Username string `json:"username"`
	Password string `json:"password"`
	IsAdmin  bool   `json:"is_admin"`
}
