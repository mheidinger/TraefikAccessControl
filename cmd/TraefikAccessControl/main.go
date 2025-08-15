package main

import (
	"flag"
	"os"
	"strconv"

	"TraefikAccessControl/manager"
	"TraefikAccessControl/repository"
	"TraefikAccessControl/server"

	log "github.com/sirupsen/logrus"
)

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

func main() {
	// Priority order for configuration: flag > env > default
	dbNameDefault := getEnv("DB_NAME", "tac.db")
	importNameDefault := getEnv("IMPORT_NAME", "")
	forceImportDefault := getEnv("FORCE_IMPORT", "false")
	cookieNameDefault := getEnv("COOKIE_NAME", "tac_token")
	userHeaderNameDefault := getEnv("USER_HEADER_NAME", "X-TAC-User")
	portDefault := getEnv("PORT", "4181")
	externalURLDefault := getEnv("EXTERNAL_URL", "")

	dbNamePtr := flag.String("db_name", dbNameDefault, "Path of the database file; Can also be set via the DB_NAME environment variable")
	importNamePtr := flag.String("import_name", importNameDefault, "Path of an file to import; Can also be set via the IMPORT_NAME environment variable")
	forceImportPtr := flag.Bool("force_import", forceImportDefault == "true", "Force the import of the given file, deletes all existing data; Can also be set via the FORCE_IMPORT environment variable")
	cookieNamePtr := flag.String("cookie_name", cookieNameDefault, "Cookie name used; Can also be set via the COOKIE_NAME environment variable")
	userHeaderNamePtr := flag.String("user_header_name", userHeaderNameDefault, "Header name that contains the username after successful auth; Can also be set via the USER_HEADER_NAME environment variable")
	portPtr := flag.Int("port", func() int { p, _ := strconv.Atoi(portDefault); return p }(), "Port on which the application will run; Can also be set via the PORT environment variable")
	externalUrlPtr := flag.String("external_url", externalURLDefault, "External URL of the application; Can also be set via the EXTERNAL_URL environment variable")
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

	authMgr := manager.CreateAuthManager(userRep, tokenRep)
	siteMgr := manager.CreateSiteManager(siteRep, siteMappingRep)
	accessMgr := manager.CreateAccessManager()
	importExportManager := manager.CreateImportExportManager()

	if importNamePtr != nil && *importNamePtr != "" {
		err = importExportManager.ImportFile(*importNamePtr, *forceImportPtr)
		if err != nil {
			log.Warn("Abort: Failed to import data")
		}
	}

	if count, err := authMgr.GetUserCount(); err == nil && count == 0 {
		createErr := authMgr.CreateFirstUser()
		if createErr != nil {
			log.WithError(createErr).Error("Failed to create first user")
		}
	}

	var externalURL *string
	if externalUrlPtr != nil && *externalUrlPtr != "" {
		externalURL = externalUrlPtr
	}
	srv := server.NewServer(*cookieNamePtr, *userHeaderNamePtr, externalURL)

	log.WithField("port", *portPtr).Info("Listening on specified port")
	log.Info(srv.Router.Run(":" + strconv.Itoa(*portPtr)))

	authMgr.Close()
	siteMgr.Close()
	accessMgr.Close()
}
