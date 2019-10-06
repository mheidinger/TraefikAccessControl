package repository

import (
	"TraefikAccessControl/models"

	log "github.com/sirupsen/logrus"
)

// Add used models to enable auto migration for them
func init() {
	databaseModels = append(databaseModels, &models.Site{})
}

type SiteRepository struct{}

func CreateSiteRepository() (*SiteRepository, error) {
	if databaseConnection == nil {
		return nil, ErrGormNotInitialized
	}
	return &SiteRepository{}, nil
}

func (rep *SiteRepository) Create(site *models.Site) (err error) {
	err = databaseConnection.Create(site).Error
	if err != nil {
		log.WithField("err", err).Error("Could not create site")
		return
	}
	return
}
