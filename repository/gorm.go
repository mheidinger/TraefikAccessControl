package repository

import (
	"errors"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/sqlite" // Needed for gorm
	log "github.com/sirupsen/logrus"
)

// ErrGormNotInitialized is returned if a repository is initialized before the database connection
var ErrGormNotInitialized = errors.New("db repository: gorm repository must be initialized first")

// databaseConnection is shared between most repositories
var (
	databaseConnection *gorm.DB
	databaseModels     []interface{} // Contains pointers to all models that should be automigrated by gorm initialization
)

// InitDatabaseConnection initializes the gorm connection
func InitDatabaseConnection(name string) error {
	db, err := gorm.Open("sqlite3", name)
	if err != nil {
		log.WithFields(log.Fields{"db_name": name, "err": err}).Error("Could not open datbase")
		return err
	}
	log.WithField("db_name", name).Info("Initialized database connection")

	err = db.AutoMigrate(databaseModels...).Error
	if err != nil {
		log.WithFields(log.Fields{"db_name": name, "err": err}).Error("Failed to auto migrate db structs")
		return err
	}
	databaseConnection = db

	return nil
}

// CloseDatabaseConnection closes the gorm connection
func CloseDatabaseConnection() {
	if err := databaseConnection.Close(); err != nil {
		log.WithField("err", err).Fatal("Error shutting down gorm")
		return
	}

	databaseConnection = nil
}

// IsRecordNotFoundError checks whether an error is 'record not found'
func IsRecordNotFoundError(err error) bool {
	return gorm.IsRecordNotFoundError(err)
}
