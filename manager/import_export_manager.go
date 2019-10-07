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
	data, err := ioutil.ReadFile(filePath)
	if err != nil {
		log.WithFields(log.Fields{"import_path": filePath, "err": err}).Error("Failed to open import file")
		return err
	}

	importExport := &models.ImportExport{}
	err = json.Unmarshal(data, importExport)
	if err != nil {
		log.WithFields(log.Fields{"import_path": filePath, "err": err}).Error("Failed to unmarshal import file")
		return err
	}

	if force {
		GetAuthManager().ClearAll()
		GetSiteManager().ClearAll()
	}

	userIDMapping := make(map[int]int, 0)
	for _, user := range importExport.Users {
		importID := user.ID
		user.ID = 0

		err = GetAuthManager().CreateUser(user)
	}

	return
}
