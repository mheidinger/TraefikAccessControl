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
		tokenLength: 30,
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

	if count, err := mgr.GetUserCount(); err == nil && count == 1 {
		createLog.Info("Make first user an admin")
		user.IsAdmin = true
		err = mgr.userRep.Update(user)
		if err != nil {
			createLog.WithField("err", err).Error("Failed to make first user an admin")
		}
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

func (mgr *AuthManager) ValidateTokenString(tokenString string, fromBearer bool) (user *models.User, err error) {
	token, err := mgr.tokenRep.GetByTokenString(tokenString)
	if err != nil && !repository.IsRecordNotFoundError(err) {
		log.WithField("err", err).Error("Failed to get token by string")
		return nil, fmt.Errorf("Failed to get token by string")
	} else if repository.IsRecordNotFoundError(err) {
		return nil, fmt.Errorf("Token not found")
	}

	if time.Now().After(token.ExpiresAt) {
		return nil, fmt.Errorf("Token expired")
	}

	if token.IsBearer != fromBearer {
		return nil, fmt.Errorf("Bearer status of token does not math auth method")
	}

	user, err = mgr.userRep.GetByID(token.UserID)
	if err != nil {
		log.WithFields(log.Fields{"userID": token.UserID, "err": err}).Error("Failed to get user of token")
		return nil, fmt.Errorf("Failed to get token user")
	}
	return
}

func (mgr *AuthManager) ValidateCredentials(username, password string) (user *models.User, err error) {
	user, err = mgr.userRep.GetByUsername(username)
	if err != nil && !repository.IsRecordNotFoundError(err) {
		log.WithField("err", err).Error("Failed to get user by username")
		return nil, fmt.Errorf("Failed to get user by username")
	} else if repository.IsRecordNotFoundError(err) {
		return nil, fmt.Errorf("User not found")
	}

	if ok, err := crypt.ValidateScryptPassword(password, user.Password); err == nil && !ok {
		return nil, fmt.Errorf("Password not correct")
	} else if err != nil {
		log.WithField("err", err).Error("Failed to validate scrypt password")
		return nil, fmt.Errorf("Failed to validate password")
	}

	return
}

func (mgr *AuthManager) DeleteToken(tokenString string) (err error) {
	return mgr.tokenRep.DeleteByTokenString(tokenString)
}

func (mgr *AuthManager) GetAllUsers() (users []*models.User, err error) {
	return mgr.userRep.GetAll()
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

func (mgr *AuthManager) GetUserCount() (count int, err error) {
	count, err = mgr.userRep.Count()
	if err != nil {
		log.WithField("err", err).Error("Failed to get user count")
		return -1, fmt.Errorf("Failed to get user count")
	}
	return
}
