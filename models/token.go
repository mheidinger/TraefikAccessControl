package models

import "time"

type Token struct {
	Token     string    `json:"token" gorm:"primary_key"`
	UserID    int       `json:"user_id"`
	ExpiresAt time.Time `json:"expires_at"`
	Name      *string   `json:"name"` // A set name indicates that this is a bearer token
	CreatedAt time.Time `json:"deletet_at"`
}

func (t *Token) IsBearer() bool {
	return t.Name != nil
}
