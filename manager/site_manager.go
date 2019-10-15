package manager

import (
	"TraefikAccessControl/models"
	"TraefikAccessControl/repository"
	"fmt"
	"strings"

	log "github.com/sirupsen/logrus"
)

type SiteManager struct {
	siteRep        *repository.SiteRepository
	siteMappingRep *repository.SiteMappingRepository
	done           chan struct{}
}

var siteManager *SiteManager

func CreateSiteManager(siteRep *repository.SiteRepository, siteMappingRep *repository.SiteMappingRepository) *SiteManager {
	if siteManager != nil {
		return siteManager
	}

	siteManager = &SiteManager{
		siteRep:        siteRep,
		siteMappingRep: siteMappingRep,
		done:           make(chan struct{}),
	}
	return siteManager
}

func GetSiteManager() *SiteManager {
	return siteManager
}

func (mgr *SiteManager) Close() {
	mgr.done <- struct{}{}
}

func (mgr *SiteManager) CreateSite(site *models.Site) (err error) {
	createLog := log.WithFields(log.Fields{"host": site.Host, "pathPrefix": site.PathPrefix})

	if site.Host == "" || site.ID != 0 {
		return fmt.Errorf("Site not valid")
	}

	site.Host = strings.TrimSpace(site.Host)
	site.PathPrefix = strings.TrimSpace(site.PathPrefix)
	if !strings.HasPrefix(site.PathPrefix, "/") {
		site.PathPrefix = "/" + site.PathPrefix
	}

	err = mgr.siteRep.Create(site)
	if err != nil {
		createLog.WithField("err", err).Error("Failed to save site")
		return fmt.Errorf("Failed to save site")
	}

	return
}

func (mgr *SiteManager) GetSite(host, path string) (site *models.Site, err error) {
	sites, err := mgr.siteRep.GetByHost(host)

	var matchLength = -1
	for _, checkSite := range sites {
		if strings.HasPrefix(path, checkSite.PathPrefix) && len(checkSite.PathPrefix) > matchLength {
			site = checkSite
		}
	}

	if site == nil {
		log.WithFields(log.Fields{"host": host, "path": path}).Warn("Matching site not found")
		return nil, fmt.Errorf("Matching site not found")
	}
	return
}

func (mgr *SiteManager) GetSiteByID(siteID int) (site *models.Site, err error) {
	return mgr.siteRep.GetByID(siteID)
}

func (mgr *SiteManager) GetAllSites() (sites []*models.Site, err error) {
	return mgr.siteRep.GetAll()
}

func (mgr *SiteManager) UpdateSite(site *models.Site) (err error) {
	if !strings.HasPrefix(site.PathPrefix, "/") {
		site.PathPrefix = "/" + site.PathPrefix
	}
	return mgr.siteRep.Update(site)
}

func (mgr *SiteManager) DeleteSite(siteID int) (err error) {
	deleteLog := log.WithField("siteID", siteID)

	err = mgr.siteMappingRep.DeleteBySite(siteID)
	if err != nil {
		deleteLog.WithField("err", err).Error("Could not delete site mapping for to be deleted site")
	}

	return mgr.siteRep.DeleteByID(siteID)
}

func (mgr *SiteManager) CreateSiteMapping(siteMapping *models.SiteMapping) (err error) {
	createLog := log.WithFields(log.Fields{"userID": siteMapping.UserID, "siteID": siteMapping.SiteID, "BasicAuthAllowed": siteMapping.BasicAuthAllowed})

	if siteMapping.UserID <= 0 || siteMapping.SiteID <= 0 {
		return fmt.Errorf("SiteMapping not valid")
	}

	err = mgr.siteMappingRep.Create(siteMapping)
	if err != nil {
		createLog.WithField("err", err).Error("Failed to save site")
		return fmt.Errorf("Failed to save site mapping")
	}

	return
}

func (mgr *SiteManager) GetSiteMapping(user *models.User, site *models.Site) (siteMapping *models.SiteMapping, err error) {
	siteMapping, err = mgr.siteMappingRep.GetByUserSite(user.ID, site.ID)
	if err != nil && !repository.IsRecordNotFoundError(err) {
		log.WithFields(log.Fields{"userID": user.ID, "siteID": site.ID, "err": err}).Error("Failed to get site mapping")
		return nil, fmt.Errorf("Failed to get site mapping")
	}
	return
}

func (mgr *SiteManager) GetSiteMappingsByUser(user *models.User) (siteMappings []*models.SiteMapping, err error) {
	return mgr.siteMappingRep.GetByUser(user.ID)
}

func (mgr *SiteManager) GetSiteMappingsBySite(site *models.Site) (siteMappings []*models.SiteMapping, err error) {
	return mgr.siteMappingRep.GetBySite(site.ID)
}

func (mgr *SiteManager) DeleteSiteMapping(userID, siteID int) (err error) {
	return mgr.siteMappingRep.Delete(userID, siteID)
}

func (mgr *SiteManager) DeleteUserMappings(userID int) (err error) {
	return mgr.siteMappingRep.DeleteByUser(userID)
}

func (mgr *SiteManager) ClearAll() (err error) {
	err = mgr.siteRep.Clear()
	if err != nil {
		return
	}
	err = mgr.siteMappingRep.Clear()
	if err != nil {
		return
	}
	return
}
