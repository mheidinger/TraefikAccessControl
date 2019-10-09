package manager

import (
	"TraefikAccessControl/models"
	"encoding/json"
	"io/ioutil"

	log "github.com/sirupsen/logrus"
)

type ImportExportManager struct {
	done chan struct{}
}

var importExportManager *ImportExportManager

func CreateImportExportManager() *ImportExportManager {
	if importExportManager != nil {
		return importExportManager
	}

	importExportManager = &ImportExportManager{
		done: make(chan struct{}),
	}
	return importExportManager
}

func GetImportExportManager() *ImportExportManager {
	return importExportManager
}

func (mgr *ImportExportManager) Close() {
	mgr.done <- struct{}{}
}

func (mgr *ImportExportManager) ImportFile(filePath string, force bool) (err error) {
	importLog := log.WithField("import_path", filePath)

	data, err := ioutil.ReadFile(filePath)
	if err != nil {
		importLog.WithField("err", err).Error("Failed to open import file")
		return err
	}

	importExport := &models.ImportExport{}
	err = json.Unmarshal(data, importExport)
	if err != nil {
		importLog.WithField("err", err).Error("Failed to unmarshal import file")
		return err
	}

	if force {
		GetAuthManager().ClearAll()
		GetSiteManager().ClearAll()
	}

	userIDMapping := make(map[int]int)
	for _, user := range importExport.Users {
		importID := user.ID
		user.ID = 0

		token, err := GetAuthManager().CreateUser(user)
		log.WithFields(log.Fields{"token": token.Token, "username": user.Username}).Info("User token")
		if err != nil {
			importLog.Warn("Failed to create user from import")
			continue
		}
		userIDMapping[importID] = user.ID
	}

	siteIDMapping := make(map[int]int)
	for _, site := range importExport.Sites {
		importID := site.ID
		site.ID = 0

		err = GetSiteManager().CreateSite(site)
		if err != nil {
			importLog.Warn("Failed to create site from import")
			continue
		}
		siteIDMapping[importID] = site.ID
	}

	for _, siteMapping := range importExport.SiteMappings {
		realUserID, ok := userIDMapping[siteMapping.UserID]
		if !ok {
			importLog.WithField("userID", siteMapping.UserID).Warn("UserID not found in mapping")
			continue
		}
		realSiteID, ok := siteIDMapping[siteMapping.SiteID]
		if !ok {
			importLog.WithField("siteID", siteMapping.SiteID).Warn("SiteID not found in mapping")
			continue
		}

		siteMapping.UserID = realUserID
		siteMapping.SiteID = realSiteID
		err = GetSiteManager().CreateSiteMapping(siteMapping)
		if err != nil {
			importLog.Warn("Failed to create site mapping from import")
			continue
		}
	}

	return
}
