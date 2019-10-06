package repository

import (
	"TraefikAccessControl/models"

	log "github.com/sirupsen/logrus"
)

// Add used models to enable auto migration for them
func init() {
	databaseModels = append(databaseModels, &models.User{})
}

type UserRepository struct{}

func CreateUserRepository() (*UserRepository, error) {
	if databaseConnection == nil {
		return nil, ErrGormNotInitialized
	}
	return &UserRepository{}, nil
}

func (rep *UserRepository) Create(user *models.User) (err error) {
	err = databaseConnection.Create(user).Error
	if err != nil {
		log.WithField("err", err).Error("Could not create user")
		return
	}
	return
}
