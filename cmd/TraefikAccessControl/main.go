package main

import (
	"flag"

	"TraefikAccessControl/manager"
	"TraefikAccessControl/repository"
	"TraefikAccessControl/server"

	log "github.com/sirupsen/logrus"
)

func main() {
	dbNamePtr := flag.String("db_name", "tac.db", "Path of the database file")
	importNamePtr := flag.String("import_name", "", "Path of an file to import")
	forceImportPtr := flag.Bool("force_import", false, "Force the import of the given file, deletes all existing data")
	flag.Parse()

	err := repository.InitDatabaseConnection(*dbNamePtr)
	if err != nil {
		log.Fatal("Abort: Failed to initialize database")
	}

	userRep, err := repository.CreateUserRepository()
	if err != nil {
		log.Fatal("Abort: Failed to create user repository")
	}
	tokenRep, err := repository.CreateTokenRepository()
	if err != nil {
		log.Fatal("Abort: Failed to create token repository")
	}
	siteRep, err := repository.CreateSiteRepository()
	if err != nil {
		log.Fatal("Abort: Failed to create site repository")
	}
	siteMappingRep, err := repository.CreateSiteMappingRepository()
	if err != nil {
		log.Fatal("Abort: Failed to create site mapping repository")
	}

	_ = manager.CreateAuthManager(userRep, tokenRep)
	_ = manager.CreateSiteManager(siteRep, siteMappingRep)
	_ = manager.CreateAccessManager()
	importExportManager := manager.CreateImportExportManager()

	if importNamePtr != nil && *importNamePtr != "" {
		err = importExportManager.ImportFile(*importNamePtr, *forceImportPtr)
		if err != nil {
			log.Warn("Abort: Failed to import data")
		}
	}

	srv := server.NewServer()

	// Start
	log.WithField("port", 4181).Info("Listening on specified port")
	log.Info(srv.Router.Run(":4181"))
}
