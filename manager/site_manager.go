package manager

import (
	"TraefikAccessControl/models"
	"TraefikAccessControl/repository"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
)

type SiteManager struct {
	siteRep             *repository.SiteRepository
	siteMappingRep      *repository.SiteMappingRepository
	checkConfigInterval int
	done                chan struct{}
}

var siteManager *SiteManager

const (
	CheckConfigHeader        = "X-TAC-Check-Config"
	CheckConfigHeaderContent = "checked"
)

func CreateSiteManager(siteRep *repository.SiteRepository, siteMappingRep *repository.SiteMappingRepository) *SiteManager {
	if siteManager != nil {
		return siteManager
	}

	siteManager = &SiteManager{
		siteRep:             siteRep,
		siteMappingRep:      siteMappingRep,
		checkConfigInterval: 5,
		done:                make(chan struct{}),
	}
	go siteManager.checkConfigRoutine()
	return siteManager
}

func (mgr *SiteManager) checkConfigRoutine() {
	log.WithField("interval", mgr.checkConfigInterval).Info("Site check will run on fixed hourly interval")
	mgr.checkAllSitesConfig()
	ticker := time.NewTicker(time.Minute * time.Duration(mgr.checkConfigInterval))
	for {
		select {
		case <-mgr.done:
			return
		case <-ticker.C:
			log.Info("Check all site configs")
			mgr.checkAllSitesConfig()
		}
	}
}

func (mgr *SiteManager) checkAllSitesConfig() {
	sites, err := mgr.siteRep.GetAll()
	if err != nil {
		log.WithField("err", err).Error("Failed to get all sites")
		return
	}
	for _, site := range sites {
		mgr.checkSiteConfig(site)
	}
}

func (mgr *SiteManager) checkSiteConfig(site *models.Site) {
	requestURL, err := url.Parse("http://" + site.Host + site.PathPrefix)
	if err != nil {
		log.WithFields(log.Fields{"err": err, "host": site.Host, "path-prefix": site.PathPrefix}).Error("Failed to parse request url from host and pathprefix")
		return
	}

	client := &http.Client{}
	req, _ := http.NewRequest("GET", requestURL.String(), nil)
	if err != nil {
		log.WithFields(log.Fields{"err": err, "url": requestURL.String()}).Error("Failed to create request")
	}
	req.Header.Add(CheckConfigHeader, "site-check")

	resp, err := client.Do(req)
	var header string
	if err != nil {
		log.WithFields(log.Fields{"err": err, "url": requestURL.String()}).Error("Failed to send request")
	} else {
		header = resp.Header.Get(CheckConfigHeader)
		resp.Body.Close()
	}

	if err == nil && header == CheckConfigHeaderContent {
		site.ConfigOK = true
	} else {
		site.ConfigOK = false
	}

	err = mgr.UpdateSite(site)
	if err != nil {
		log.WithField("err", err).Error("Failed to update config OK status")
		return
	}
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

	mgr.checkSiteConfig(site)

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

func (mgr *SiteManager) GetAnonymousSites() (sites []*models.Site, err error) {
	return mgr.siteRep.GetAnonymous()
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
