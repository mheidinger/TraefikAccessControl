package main

import (
	"net/http"

	"TraefikAccessControl/manager"
	"TraefikAccessControl/repository"
	"TraefikAccessControl/server"

	log "github.com/sirupsen/logrus"
)

func main() {
	err := repository.InitDatabaseConnection("tac.db")
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

	manager.InitAuthManager(userRep, tokenRep)
	manager.InitSiteManager(siteRep)

	srv := server.NewServer()

	// Start
	log.Info("Listening on :4181")
	log.Info(http.ListenAndServe(":4181", srv.Router))
}
