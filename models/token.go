package models

import "time"

type Token struct {
	token     string
	expiresAt time.Time
	isBearer  bool
}
