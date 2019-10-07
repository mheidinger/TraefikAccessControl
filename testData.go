package main

import (
	"TraefikAccessControl/models"
	"encoding/json"
	"io/ioutil"
)

func main() {
	importExport := models.ImportExport{}

	importExport.Users = append(importExport.Users, &models.User{ID: 1, Username: "mheidinger", Password: "123456", IsAdmin: true})
	importExport.Users = append(importExport.Users, &models.User{ID: 2, Username: "test", Password: "654321", IsAdmin: false})

	importExport.Sites = append(importExport.Sites, &models.Site{ID: 1, Host: "example.com", PathPrefix: ""})
	importExport.Sites = append(importExport.Sites, &models.Site{ID: 2, Host: "noexample.de", PathPrefix: "/restricted"})

	importExport.SiteMappings = append(importExport.SiteMappings, &models.SiteMapping{UserID: 1, SiteID: 1, BasicAuthAllowed: false})
	importExport.SiteMappings = append(importExport.SiteMappings, &models.SiteMapping{UserID: 2, SiteID: 2, BasicAuthAllowed: true})

	jsonB, _ := json.MarshalIndent(importExport, "", "  ")

	ioutil.WriteFile("tac_data.json", jsonB, 0644)
}
