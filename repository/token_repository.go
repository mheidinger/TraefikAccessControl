package repository

import (
	"TraefikAccessControl/models"

	log "github.com/sirupsen/logrus"
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

func (rep *TokenRepository) GetByTokenString(tokenString string) (token *models.Token, err error) {
	token = &models.Token{}
	err = databaseConnection.First(token, &models.Token{Token: tokenString}).Error
	return
}

func (rep *TokenRepository) GetBearerByUser(userID int) (tokens []*models.Token, err error) {
	err = databaseConnection.Where("user_id = ? AND name NOT NULL", userID).Order("name").Find(&tokens).Error
	return
}

func (rep *TokenRepository) DeleteByTokenString(tokenString string) (err error) {
	err = databaseConnection.Delete(&models.Token{Token: tokenString}).Error
	return
}

func (rep *TokenRepository) DeleteByUser(userID int) (err error) {
	err = databaseConnection.Where(&models.Token{UserID: userID}).Delete(&models.User{}).Error
	return
}

func (rep *TokenRepository) DeleteByUserName(userID int, name string) (err error) {
	log.Info(name)
	err = databaseConnection.Where(&models.Token{UserID: userID, Name: &name}).Delete(&models.Token{}).Error
	return
}

func (rep *TokenRepository) Clear() (err error) {
	err = databaseConnection.Delete(&models.Token{}).Error
	return
}
