package models

import "time"

type Token struct {
	Token     string    `json:"token" gorm:"primary_key"`
	ExpiresAt time.Time `json:"expires_at"`
	IsBearer  bool      `json:"is_bearer"`
}
