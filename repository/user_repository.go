package repository

import (
	"TraefikAccessControl/models"
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
	return
}

func (rep *UserRepository) GetByID(id int) (user *models.User, err error) {
	user = &models.User{}
	err = databaseConnection.First(user, &models.User{ID: id}).Error
	return
}

func (rep *UserRepository) GetByUsername(username string) (user *models.User, err error) {
	user = &models.User{}
	err = databaseConnection.First(user, &models.User{Username: username}).Error
	return
}

func (rep *UserRepository) GetAll() (users []*models.User, err error) {
	err = databaseConnection.Find(&users).Error
	return
}

func (rep *UserRepository) Update(user *models.User) (err error) {
	err = databaseConnection.Save(user).Error
	return
}

func (rep *UserRepository) Clear() (err error) {
	err = databaseConnection.Delete(&models.User{}).Error
	return
}

func (rep *UserRepository) Count() (count int, err error) {
	err = databaseConnection.Model(&models.User{}).Count(&count).Error
	return
}
