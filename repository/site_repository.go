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
	err = databaseConnection.Where(&models.Site{Host: host}).Order("path_prefix").Find(&sites).Error
	return
}

func (rep *SiteRepository) GetByID(siteID int) (site *models.Site, err error) {
	site = &models.Site{}
	err = databaseConnection.First(site, &models.Site{ID: siteID}).Error
	return
}

func (rep *SiteRepository) GetAnonymous() (sites []*models.Site, err error) {
	err = databaseConnection.Where(&models.Site{AnonymousAccess: true}).Order("host, path_prefix").Find(&sites).Error
	return
}

func (rep *SiteRepository) GetAll() (sites []*models.Site, err error) {
	err = databaseConnection.Order("host, path_prefix").Find(&sites).Error
	return
}

func (rep *SiteRepository) DeleteByID(siteID int) (err error) {
	err = databaseConnection.Where(&models.Site{ID: siteID}).Delete(&models.Site{}).Error
	return
}

func (rep *SiteRepository) Update(site *models.Site) (err error) {
	err = databaseConnection.Save(site).Error
	return
}

func (rep *SiteRepository) Clear() (err error) {
	err = databaseConnection.Delete(&models.Site{}).Error
	return
}
