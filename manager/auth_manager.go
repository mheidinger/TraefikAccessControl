package manager

import (
	"TraefikAccessControl/models"
	"TraefikAccessControl/repository"
	"TraefikAccessControl/utils"
	"TraefikAccessControl/utils/crypt"
	"fmt"
	"time"

	log "github.com/sirupsen/logrus"
)

type AuthManager struct {
	userRep     *repository.UserRepository
	tokenRep    *repository.TokenRepository
	tokenLength int
	tokenExpiry int
	done        chan struct{}
}

var authManager *AuthManager

func CreateAuthManager(userRep *repository.UserRepository, tokenRep *repository.TokenRepository) *AuthManager {
	if authManager != nil {
		return authManager
	}

	authManager = &AuthManager{
		userRep:     userRep,
		tokenRep:    tokenRep,
		tokenLength: 20,
		tokenExpiry: 24,
		done:        make(chan struct{}),
	}
	return authManager
}

func GetAuthManager() *AuthManager {
	return authManager
}

func (mgr *AuthManager) Close() {
	mgr.done <- struct{}{}
}

func (mgr *AuthManager) CreateUser(user *models.User) (*models.Token, error) {
	createLog := log.WithFields(log.Fields{"username": user.Username})

	if user.Username == "" || user.Password == "" || user.ID != 0 {
		return nil, fmt.Errorf("User not valid")
	}

	existingUser, err := mgr.userRep.GetByUsername(user.Username)
	if err != nil && !repository.IsRecordNotFoundError(err) {
		createLog.WithField("err", err).Warn("Could not validate whether user already exists")
	} else if err == nil && existingUser != nil && existingUser.ID > 0 {
		return nil, fmt.Errorf("User already exists")
	}

	user.Password, err = crypt.HashScrypt(user.Password)
	if err != nil {
		createLog.WithField("err", err).Error("Password hashing failed")
		return nil, fmt.Errorf("Password Hashing failed")
	}
	user.IsAdmin = false

	err = mgr.userRep.Create(user)
	if err != nil {
		createLog.WithField("err", err).Error("Failed to save new user")
		return nil, fmt.Errorf("Saving user failed")
	}

	return mgr.CreateUserToken(user.ID, false)
}

func (mgr *AuthManager) CreateUserToken(userID int, isBearer bool) (token *models.Token, err error) {
	createLogger := log.WithFields(log.Fields{"userID": userID, "isBearer": isBearer})

	token = &models.Token{
		UserID:   userID,
		Token:    utils.RandomString(mgr.tokenLength),
		IsBearer: isBearer,
	}

	if !isBearer {
		token.ExpiresAt = time.Now().UTC().Add(time.Hour * time.Duration(mgr.tokenExpiry))
	}

	err = mgr.tokenRep.Create(token)
	if err != nil {
		createLogger.WithField("err", err).Error("Failed to save token")
		return nil, fmt.Errorf("Failed to save token")
	}

	return
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
