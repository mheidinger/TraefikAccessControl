package manager

import "TraefikAccessControl/repository"

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
