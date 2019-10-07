package manager

import "TraefikAccessControl/repository"

type AuthManager struct {
	userRep  *repository.UserRepository
	tokenRep *repository.TokenRepository
	done     chan struct{}
}

var authManager *AuthManager

func CreateAuthManager(userRep *repository.UserRepository, tokenRep *repository.TokenRepository) *AuthManager {
	if authManager != nil {
		return authManager
	}

	authManager = &AuthManager{
		done: make(chan struct{}),
	}
	return authManager
}

func GetAuthManager() *AuthManager {
	return authManager
}

func (mgr *AuthManager) Close() {
	mgr.done <- struct{}{}
}

func (mgr *AuthManager) CreateUser() {

}

func (mgr *AuthManager) ClearAll() (err error) {
	err = mgr.userRep.Clear()
	if err != nil {
		return
	}
	err = mgr.tokenRep.Clear()
	if err != nil {
		return
	}
	return
}
