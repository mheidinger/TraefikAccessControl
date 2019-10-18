package manager

import (
	"TraefikAccessControl/models"
	"TraefikAccessControl/repository"
	"fmt"

	log "github.com/sirupsen/logrus"
)

type AccessManager struct {
	done chan struct{}
}

var accessManager *AccessManager

func CreateAccessManager() *AccessManager {
	if accessManager != nil {
		return accessManager
	}

	accessManager = &AccessManager{
		done: make(chan struct{}),
	}
	return accessManager
}

func GetAccessManager() *AccessManager {
	return accessManager
}

func (mgr *AccessManager) Close() {
	mgr.done <- struct{}{}
}

func (mgr *AccessManager) CheckAccessToken(host, path, tokenString string, fromBearer bool, requestLogger *log.Entry) (ok bool, user *models.User, err error) {
	user, err = GetAuthManager().ValidateTokenString(tokenString, fromBearer)
	if err != nil {
		requestLogger.WithField("err", err).Info("Token not valid")
		return
	}
	ok, err = mgr.CheckAccess(host, path, user, false, requestLogger)
	return
}

func (mgr *AccessManager) CheckAccessCredentials(host, path, username, password string, fromBasicAuth bool, requestLogger *log.Entry) (ok bool, user *models.User, err error) {
	user, err = GetAuthManager().ValidateCredentials(username, password)
	if err != nil {
		requestLogger.WithField("err", err).Info("Credentials not valid")
		return
	}
	ok, err = mgr.CheckAccess(host, path, user, fromBasicAuth, requestLogger)
	return
}

func (mgr *AccessManager) CheckAnonymousAccess(host, path string, requestLogger *log.Entry) (ok bool, err error) {
	site, err := GetSiteManager().GetSite(host, path)
	if err != nil {
		requestLogger.WithField("err", err).Info("Site not found")
		return
	}

	if site.AnonymousAccess {
		return true, nil
	}
	return false, nil
}

func (mgr *AccessManager) CheckAccess(host, path string, user *models.User, fromBasicAuth bool, requestLogger *log.Entry) (ok bool, err error) {
	site, err := GetSiteManager().GetSite(host, path)
	if err != nil {
		requestLogger.WithField("err", err).Info("Site not found")
		return
	}

	if site.AnonymousAccess {
		return true, nil
	}

	siteMapping, err := GetSiteManager().GetSiteMapping(user, site)
	if err != nil && !repository.IsRecordNotFoundError(err) {
		requestLogger.WithField("err", err).Error("Failed to get site mapping")
		return
	} else if repository.IsRecordNotFoundError(err) {
		requestLogger.Info("User does not have access to this site")
		return false, fmt.Errorf("User does not have access to this site")
	}

	if !siteMapping.BasicAuthAllowed && fromBasicAuth {
		requestLogger.Info("BasicAuth not allowed for this user and site")
		return false, fmt.Errorf("BasicAuth not allowed for this user and site")
	}

	return true, nil
}
