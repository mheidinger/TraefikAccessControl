package models

import "time"

type Token struct {
	Token     string    `json:"token,omitempty" gorm:"primary_key"`
	ExpiresAt time.Time `json:"expires_at,omitempty"`
	IsBearer  bool      `json:"is_bearer,omitempty"`
}
