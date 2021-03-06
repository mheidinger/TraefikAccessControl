package models

type Site struct {
	ID              int    `json:"id" gorm:"primary_key;auto_increment"`
	Host            string `json:"host"`
	PathPrefix      string `json:"path_prefix"`
	AnonymousAccess bool   `json:"anonymous_access"`
	PromptBasicAuth bool   `json:"prompt_basic_auth"`
	ConfigOK        bool   `json:"config_ok"`
}
