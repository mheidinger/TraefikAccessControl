package models

type Site struct {
	ID         int    `json:"id,omitempty" gorm:"primary_key;auto_increment"`
	Host       string `json:"host,omitempty"`
	PathPrefix string `json:"path_prefix,omitempty"`
}
