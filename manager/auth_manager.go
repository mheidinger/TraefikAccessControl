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
	userRep              *repository.UserRepository
	tokenRep             *repository.TokenRepository
	tokenLength          int
	tokenExpiry          int
	tokenCleanupInterval int
	done                 chan struct{}
}

var authManager *AuthManager

func CreateAuthManager(userRep *repository.UserRepository, tokenRep *repository.TokenRepository) *AuthManager {
	if authManager != nil {
		return authManager
	}

	authManager = &AuthManager{
		userRep:              userRep,
		tokenRep:             tokenRep,
		tokenLength:          30,
		tokenExpiry:          24,
		tokenCleanupInterval: 2,
		done:                 make(chan struct{}),
	}
	go authManager.cleanupExpiredSessionsRoutine()
	return authManager
}

func (mgr *AuthManager) cleanupExpiredSessionsRoutine() {
	log.WithField("interval", mgr.tokenCleanupInterval).Info("Session cleaner will run on fixed hourly interval")
	mgr.tokenRep.DeleteExpired()
	ticker := time.NewTicker(time.Hour * time.Duration(mgr.tokenCleanupInterval))
	for {
		select {
		case <-mgr.done:
			return
		case <-ticker.C:
			log.Info("Cleaning expired sessions")
			mgr.tokenRep.DeleteExpired()
		}
	}
}

func GetAuthManager() *AuthManager {
	return authManager
}

func (mgr *AuthManager) Close() {
	mgr.done <- struct{}{}
}

func (mgr *AuthManager) CreateUser(user *models.User) (err error) {
	createLog := log.WithFields(log.Fields{"username": user.Username})

	if user.Username == "" || user.Password == "" || user.ID != 0 {
		return fmt.Errorf("User not valid")
	}

	existingUser, err := mgr.userRep.GetByUsername(user.Username)
	if err != nil && !repository.IsRecordNotFoundError(err) {
		createLog.WithField("err", err).Warn("Could not validate whether user already exists")
	} else if err == nil && existingUser != nil && existingUser.ID > 0 {
		return fmt.Errorf("User already exists")
	}

	user.Password, err = crypt.HashScrypt(user.Password)
	if err != nil {
		createLog.WithField("err", err).Error("Password hashing failed")
		return fmt.Errorf("Password Hashing failed")
	}

	err = mgr.userRep.Create(user)
	if err != nil {
		createLog.WithField("err", err).Error("Failed to save new user")
		return fmt.Errorf("Saving user failed")
	}

	return
}

func (mgr *AuthManager) UpdateUserPassword(user *models.User, password string) (err error) {
	if password == "" {
		return fmt.Errorf("Password not valid")
	}

	user.Password, err = crypt.HashScrypt(password)
	if err != nil {
		log.WithField("err", err).Error("Failed to change user password")
		return fmt.Errorf("Saving user failed")
	}

	return mgr.userRep.Update(user)
}

func (mgr *AuthManager) DeleteUser(userID int) (err error) {
	deleteLogger := log.WithFields(log.Fields{"userID": userID})

	err = mgr.userRep.Delete(userID)
	if err != nil {
		deleteLogger.WithField("err", err).Error("Failed to delete user")
	}

	err = mgr.tokenRep.DeleteByUser(userID)
	if err != nil {
		deleteLogger.WithField("err", err).Error("Failed to delete user tokens")
	}
	return
}

func (mgr *AuthManager) CreateUserToken(userID int, name *string) (token *models.Token, err error) {
	createLogger := log.WithFields(log.Fields{"userID": userID, "name": name})

	token = &models.Token{
		UserID: userID,
		Token:  utils.RandomString(mgr.tokenLength),
		Name:   name,
	}

	if name == nil {
		token.ExpiresAt = time.Now().UTC().Add(time.Hour * time.Duration(mgr.tokenExpiry))
	}

	err = mgr.tokenRep.Create(token)
	if err != nil {
		createLogger.WithField("err", err).Error("Failed to save token")
		return nil, fmt.Errorf("Failed to save token")
	}

	return
}

func (mgr *AuthManager) GetUserByID(userID int) (user *models.User, err error) {
	return mgr.userRep.GetByID(userID)
}

func (mgr *AuthManager) GetBearerTokens(user *models.User) (tokens []*models.Token, err error) {
	return mgr.tokenRep.GetBearerByUser(user.ID)
}

func (mgr *AuthManager) ValidateTokenString(tokenString string, fromBearer bool) (user *models.User, err error) {
	token, err := mgr.tokenRep.GetByTokenString(tokenString)
	if err != nil && !repository.IsRecordNotFoundError(err) {
		log.WithField("err", err).Error("Failed to get token by string")
		return nil, fmt.Errorf("Failed to get token by string")
	} else if repository.IsRecordNotFoundError(err) {
		return nil, fmt.Errorf("Token not found")
	}

	if !token.IsBearer() && time.Now().After(token.ExpiresAt) {
		return nil, fmt.Errorf("Token expired")
	}

	if token.IsBearer() != fromBearer {
		return nil, fmt.Errorf("Bearer status of token does not match auth method")
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

func (mgr *AuthManager) CreateFirstUser() (err error) {
	user := &models.User{
		Username: "admin",
		Password: utils.RandomString(15),
		IsAdmin:  true,
	}
	log.WithFields(log.Fields{"username": user.Username, "password": user.Password}).Info("Create first user")
	err = mgr.CreateUser(user)
	return
}

func (mgr *AuthManager) DeleteToken(tokenString string) (err error) {
	return mgr.tokenRep.DeleteByTokenString(tokenString)
}

func (mgr *AuthManager) DeleteTokenByName(userID int, name string) (err error) {
	return mgr.tokenRep.DeleteByUserName(userID, name)
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
