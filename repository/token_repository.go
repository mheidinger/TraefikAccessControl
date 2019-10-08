package repository

import (
	"TraefikAccessControl/models"
)

// Add used models to enable auto migration for them
func init() {
	databaseModels = append(databaseModels, &models.Token{})
}

type TokenRepository struct{}

func CreateTokenRepository() (*TokenRepository, error) {
	if databaseConnection == nil {
		return nil, ErrGormNotInitialized
	}
	return &TokenRepository{}, nil
}

func (rep *TokenRepository) Create(token *models.Token) (err error) {
	err = databaseConnection.Create(token).Error
	return
}

func (rep *TokenRepository) Clear() (err error) {
	err = databaseConnection.Delete(&models.Token{}).Error
	return
}
