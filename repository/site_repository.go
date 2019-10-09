package repository

import (
	"TraefikAccessControl/models"
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
	return
}

func (rep *SiteRepository) GetByHost(host string) (sites []*models.Site, err error) {
	err = databaseConnection.Where(&models.Site{Host: host}).Find(&sites).Error
	return
}

func (rep *SiteRepository) Clear() (err error) {
	err = databaseConnection.Delete(&models.Site{}).Error
	return
}
