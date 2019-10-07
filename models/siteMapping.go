package models

type SiteMapping struct {
	UserID           int  `json:"user_id"`
	SiteID           int  `json:"site_id"`
	BasicAuthAllowed bool `json:"basic_auth_allowed"`
}
