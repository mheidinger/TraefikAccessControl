package models

type User struct {
	ID               int
	Username         string
	Password         string
	IsAdmin          bool
	BasicAuthAllowed bool
}
