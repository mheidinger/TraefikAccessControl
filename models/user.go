package models

type User struct {
	ID               int    `json:"id,omitempty" gorm:"primary_key;auto_increment"`
	Username         string `json:"username,omitempty"`
	Password         string `json:"password,omitempty"`
	IsAdmin          bool   `json:"is_admin,omitempty"`
	BasicAuthAllowed bool   `json:"basic_auth_allowed,omitempty"`
}
