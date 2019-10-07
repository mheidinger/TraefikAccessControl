package models

type ImportExport struct {
	Users        []*User
	Sites        []*Site
	SiteMappings []*SiteMapping
}
