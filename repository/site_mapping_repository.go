package repository

import (
	"TraefikAccessControl/models"
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
	return
}

func (rep *SiteMappingRepository) GetByUserSite(userID, siteID int) (siteMapping *models.SiteMapping, err error) {
	siteMapping = &models.SiteMapping{}
	err = databaseConnection.First(siteMapping, &models.SiteMapping{UserID: userID, SiteID: siteID}).Error
	return
}

func (rep *SiteMappingRepository) GetByUser(userID int) (siteMappings []*models.SiteMapping, err error) {
	err = databaseConnection.Where(&models.SiteMapping{UserID: userID}).Find(&siteMappings).Error
	return
}

func (rep *SiteMappingRepository) DeleteByUser(userID int) (err error) {
	err = databaseConnection.Where(&models.SiteMapping{UserID: userID}).Delete(&models.SiteMapping{}).Error
	return
}

func (rep *SiteMappingRepository) DeleteBySite(siteID int) (err error) {
	err = databaseConnection.Where(&models.SiteMapping{SiteID: siteID}).Delete(&models.SiteMapping{}).Error
	return
}

func (rep *SiteMappingRepository) Clear() (err error) {
	err = databaseConnection.Delete(&models.SiteMapping{}).Error
	return
}
