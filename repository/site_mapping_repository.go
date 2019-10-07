package repository

import (
	"TraefikAccessControl/models"

	log "github.com/sirupsen/logrus"
)

// Add used models to enable auto migration for them
func init() {
	databaseModels = append(databaseModels, &models.SiteMapping{})
}

type SiteMappingRepository struct{}

func CreateSiteMappingRepository() (*SiteMappingRepository, error) {
	if databaseConnection == nil {
		return nil, ErrGormNotInitialized
	}
	return &SiteMappingRepository{}, nil
}

func (rep *SiteMappingRepository) Create(siteMapping *models.SiteMapping) (err error) {
	err = databaseConnection.Create(siteMapping).Error
	if err != nil {
		log.WithField("err", err).Error("Could not create site mapping")
		return
	}
	return
}

func (rep *SiteMappingRepository) Clear() (err error) {
	err = databaseConnection.Delete(&models.SiteMapping{}).Error
	if err != nil {
		log.WithField("err", err).Error("Could not clear site mappings")
		return
	}
	return
}
